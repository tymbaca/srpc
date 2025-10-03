package srpc

import (
	"context"
	"errors"
)

type Connector interface {
	Connect(ctx context.Context, addr string) (ClientConn, error)
}

type ClientConn interface {
	Send(ctx context.Context, req Request) (Response, error)

	// Close must be called after Send
	Close() error
}

var ErrListenerClosed = errors.New("listener is closed")

type Listener interface {
	// Accept waits and returns new connection to the listener.
	// If Listener got closed Accept must return [ErrListenerClosed],
	// including Accept calls that didn't returned yet.
	Accept() (ServerConn, error)

	// Close closes the listener.
	// Any blocked Accept operations will be unblocked and return errors.
	// Close can be called multiple times.
	Close() error
}

type ServerConn interface {
	Request() Request
	Addr() string
	Send(ctx context.Context, resp Response) error

	// Close must be called after Send
	Close() error
}
