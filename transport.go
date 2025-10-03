package srpc

import "context"

type Connector interface {
	Connect(ctx context.Context, addr string) (ClientConn, error)
}

type ClientConn interface {
	Send(ctx context.Context, req Request) (Response, error)

	// Close must be called after Send
	Close() error
}

type Listener interface {
	// Accept waits for and returns the next connection to the listener.
	Accept() (ServerConn, error)

	// Close closes the listener.
	// Any blocked Accept operations will be unblocked and return errors.
	Close() error
}

type ServerConn interface {
	Request() Request
	Addr() string
	Send(ctx context.Context, resp Response) error

	// Close must be called after Send
	Close() error
}
