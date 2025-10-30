package srpc

import (
	"context"
	"errors"
)

type Connector interface {
	Connect(ctx context.Context, addr string) (Conn, error)
}

var ErrListenerClosed = errors.New("listener is closed")

type Listener interface {
	// Accept waits and returns new connection to the listener.
	// If Listener got closed Accept must return [ErrListenerClosed],
	// including Accept calls that didn't returned yet.
	Accept() (Conn, error)

	// Close closes the listener.
	// Any blocked Accept operations will be unblocked and return errors.
	// Close can be called multiple times.
	Close() error
}

type Conn interface {
	Addr() string
	// Can return [io.EOF]
	Read(p []byte) (n int, err error)
	Write(p []byte) (n int, err error)
	// Close must be called after Send
	Close() error
}
