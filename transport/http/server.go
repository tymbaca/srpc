package httptransport

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/tymbaca/srpc"
	"github.com/tymbaca/srpc/pkg/atomic"
)

type ServerListener struct {
	server    http.Server
	serverErr atomic.Value[error]

	conns chan srpc.ServerConn
}

func CreateAndStartListener(addr string, path string, method string) *ServerListener {
	l := NewServerListener(addr, path, method)
	go l.Start()
	return l
}

// FIX: disallow methods without body

func NewServerListener(addr string, path string, method string) *ServerListener {
	l := &ServerListener{
		server: http.Server{Addr: addr},
	}

	mux := http.NewServeMux()
	mux.HandleFunc(fmt.Sprintf("%s %s", method, path), l.handler)
	l.server.Handler = mux

	return l
}

func (l *ServerListener) Start() {
	l.conns = make(chan srpc.ServerConn)
	defer close(l.conns)

	err := l.server.ListenAndServe()
	if err != nil {
		l.serverErr.Store(err)
		return
	}
}

func (l *ServerListener) handler(w http.ResponseWriter, r *http.Request) {
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

	// FIX: may be closed
	l.conns <- conn
	<-conn.closeHandlerCh // wait until srpc handles the connection and calls [serverConn.Close]
}

// Accept waits for and returns the next connection to the listener.
func (l *ServerListener) Accept() (srpc.ServerConn, error) {
	conn, open := <-l.conns
	if !open {
		if err := l.serverErr.Load(); err != nil {
			return nil, err
		}

		return nil, errors.New("listener is closed")
	}

	return conn, nil
}

// Close closes the listener.
// Any blocked Accept operations will be unblocked and return errors.
func (l *ServerListener) Close() error {
	return l.server.Close()
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
