package srpc

import (
	"context"
	"errors"
)

// Connector connectes to another peer by it's address. Used by client.
type Connector interface {
	Connect(ctx context.Context, addr string) (Conn, error)
}

// ErrListenerClosed returned by [Listener.Accept] when listener is closed.
var ErrListenerClosed = errors.New("listener is closed")

// Listener accepts incoming connections.
//
// Multiple goroutines may invoke methods on a Listener simultaneously.
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

// Conn provides a way for peers to write and read messages (request and responses).
// Conn mimics [net.Conn]. For details see its documentation.
type Conn interface {
	RemoteAddr() string
	LocalAddr() string

	Read(p []byte) (n int, err error)
	Write(p []byte) (n int, err error)

	Close() error
}
