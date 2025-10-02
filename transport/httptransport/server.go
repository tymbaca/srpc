package httptransport

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/tymbaca/srpc"
	"github.com/tymbaca/srpc/pkg/atomic"
)

type ServerListener struct {
	server    http.Server
	serverErr atomic.Value[error]

	conns chan srpc.ServerConn
}

func NewServerListener(addr string, path string, method string) *ServerListener {
	l := &ServerListener{
		server: http.Server{Addr: addr},
		conns:  make(chan srpc.ServerConn),
	}

	mux := http.NewServeMux()
	mux.HandleFunc(fmt.Sprintf("%s %s", method, path), l.handler)
	l.server.Handler = mux

	return l
}

func (l *ServerListener) Start() {
	err := l.server.ListenAndServe()
	if err != nil {
		l.serverErr.Store(err)
		return
	}
}

func (l *ServerListener) handler(w http.ResponseWriter, r *http.Request) {
	conn := &serverConn{
		w: w, r: r,
		close: make(chan struct{}),
	}

	l.conns <- conn
	<-conn.close
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

	close chan struct{}
}

func (c *serverConn) Request() srpc.Request {
	panic("not implemented") // TODO: Implement
}

func (c *serverConn) Addr() string {
	panic("not implemented") // TODO: Implement
}

func (c *serverConn) Send(ctx context.Context, resp srpc.Response) error {
	panic("not implemented") // TODO: Implement
}

// Close must be called after Send
func (c *serverConn) Close() error {
	close(c.close)
	return nil
}
