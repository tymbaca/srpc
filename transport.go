package srpc

import "context"

type ClientTransport interface {
	Send(ctx context.Context, addr string, req Request) (Response, error)
}

// TODO: remove
// type ServerTransport interface {
// 	Listen(ctx context.Context, addr string) (ServerListener, error)
// }

type ServerListener interface {
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
}
