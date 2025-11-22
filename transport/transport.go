// Package transport provides abstract data transport layer for srpc.
package transport

import (
	"context"
	"errors"
)

// Dialer connectes to another peer by it's address. Used by client.
type Dialer interface {
	Dial(ctx context.Context, addr string) (Conn, error)
}

// ErrListenerClosed returned by [Listener.Accept] when listener is closed.
var ErrListenerClosed = errors.New("listener is closed")

// ErrListenerBadClose can be returned by [Listener.Accept] when listener closed with error.
var ErrListenerBadClose = errors.New("listener is closed with error")

// Listener accepts incoming connections.
//
// Multiple goroutines may invoke methods on a Listener simultaneously.
type Listener interface {
	// Accept waits and returns new connection to the listener.
	// If Listener got closed with [Listener.Close] Accept must
	// return [ErrListenerClosed]. including Accept calls that
	// didn't returned yet. The [Server.Start] will exit with nil after that.
	//
	// Implementation can also return [ErrListenerBadClose]. In that case
	// [Server.Start] will exit with provided error (including other errors
	// that wrap it).
	//
	// Other returned errors will be logged by the server
	// and then it will call Accept again.
	Accept() (Conn, error)

	// Close closes the listener.
	// Any blocked Accept operations will be unblocked and return errors.
	// Close can be called multiple times.
	Close() error

	// Addr returns listener's address. Address is valid to use in [Dialer.Dial].
	Addr() string
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
