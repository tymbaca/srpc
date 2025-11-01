package srpc

import (
	"context"
	"errors"
	"net"
	"time"
)

// Connector connectes to another peer by it's address. Used by client.
type Connector interface {
	net.Conn
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
// Close is called after writes are done. After that, reading conn will receive [io.EOF].
// Close can be called multiple times. After invoking Close peer still can be
// able to use Read.
type Conn interface {
	// Addr retuns address of the peer that is connected to current peer.
	// Must be valid to use in [Connector.Connect].
	Addr() string

	Read(p []byte) (n int, err error)
	Write(p []byte) (n int, err error)

	Close() error
	CloseRead() error
	CloseWrite() error

	SetDeadline(t time.Time) error
}
