package srpc

import (
	"context"
	"io"
)

type ClientTransport interface {
	Dial(ctx context.Context, addr string) (ClientConn, error)
}

type ClientConn interface {
	WriteRequest(header RequestHeader, body io.Reader) error
	ResponseHeader() (ResponseHeader, error)
	ResponseBody() (io.Reader, error)

	Close() error
}

type ServerTransport interface {
	Listen(ctx context.Context, addr string) (ClientConn, error)
}

type ServerListener interface {
	// Accept waits for and returns the next connection to the listener.
	Accept() (ServerConn, error)

	// Close closes the listener.
	// Any blocked Accept operations will be unblocked and return errors.
	Close() error
}

type ServerConn interface {
	RequestHeader() (RequestHeader, error)
	RequestBody() (io.Reader, error)
	WriteResponse(header ResponseHeader, body io.Reader) error

	// Close can be called multiple times and must be idempotent.
	Close() error
}
