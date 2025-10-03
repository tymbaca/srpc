package httptransport

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"

	"github.com/tymbaca/srpc"
	"github.com/tymbaca/srpc/pkg/atomic"
)

var ErrListenerClosed = errors.New("listener is closed")

type Listener struct {
	server    http.Server
	serverErr atomic.Value[error]

	ctx       context.Context
	ctxCancel context.CancelFunc
	closeOnce sync.Once
	conns     chan srpc.ServerConn
}

func CreateAndStartListener(addr string, path string, method string) *Listener {
	l := NewServerListener(addr, path, method)
	go l.Start()
	return l
}

// FIX: disallow methods without body

func NewServerListener(addr string, path string, method string) *Listener {
	l := &Listener{
		server: http.Server{Addr: addr},
	}

	mux := http.NewServeMux()
	mux.HandleFunc(fmt.Sprintf("%s %s", method, path), l.handler)
	l.server.Handler = mux

	return l
}

func (l *Listener) Start() {
	l.conns = make(chan srpc.ServerConn)
	l.ctx, l.ctxCancel = context.WithCancel(context.Background())
	defer l.Close()

	err := l.server.ListenAndServe()
	if err != nil {
		l.serverErr.Store(err)
		return
	}
}

// Close closes the listener.
// Any blocked Accept operations will be unblocked and return errors.
// Close can be called multiple times.
func (l *Listener) Close() (err error) {
	l.closeOnce.Do(func() { err = l.close() })
	return err
}

func (l *Listener) close() error {
	l.ctxCancel()
	return l.server.Close()
}

// Accept waits and returns new connection to the listener.
func (l *Listener) Accept() (srpc.ServerConn, error) {
	select {
	case conn, open := <-l.conns:
		if !open {
			log.Panicf("http listener: l.conns was closed, but it must not happen")
		}
		return conn, nil
	case <-l.ctx.Done():
		if err := l.serverErr.Load(); err != nil {
			return nil, err
		}

		return nil, ErrListenerClosed
	}
}

func (l *Listener) handler(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if r.Body != nil {
			r.Body.Close()
		}
	}()
	serviceMethod, metadata, err := fromHeader(r.Header)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	req := srpc.Request{
		ServiceMethod: serviceMethod,
		Metadata:      metadata,
		Body:          r.Body,
	}

	conn := &serverConn{
		w: w, r: r,
		req:            req,
		closeHandlerCh: make(chan struct{}),
	}

	// pass connection to Accept()
	select {
	case l.conns <- conn:
	case <-l.ctx.Done():
	}

	// wait until srpc handles the connection and calls [serverConn.Close]
	select {
	case <-conn.closeHandlerCh:
	case <-l.ctx.Done():
	}
}

type serverConn struct {
	w http.ResponseWriter
	r *http.Request

	req srpc.Request

	closeHandlerCh chan struct{}
}

func (c *serverConn) Request() srpc.Request {
	return c.req
}

func (c *serverConn) Addr() string {
	return c.r.RemoteAddr
}

func (c *serverConn) Send(ctx context.Context, resp srpc.Response) error {
	header, err := toHeader(resp.ServiceMethod, resp.Metadata)
	if err != nil {
		return fmt.Errorf("encode resp header: %w", err)
	}

	for k, vs := range header {
		for _, v := range vs {
			c.w.Header().Add(k, v)
		}
	}

	setStatus(c.w.Header(), resp.StatusCode)
	if resp.Error != nil {
		setError(c.w.Header())
		c.w.Write([]byte(resp.Error.Error()))
		return nil
	}

	n, err := io.Copy(c.w, resp.Body)
	if err != nil {
		return fmt.Errorf("failed to send body (%d bytes written): %w", n, err)
	}

	return nil
}

// Close must be called after Send
func (c *serverConn) Close() error {
	close(c.closeHandlerCh)
	return nil
}
