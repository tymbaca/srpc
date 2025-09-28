package srpc

import (
	"context"
	"io"
)

type ClientTransport interface {
	WriteRequest(ctx context.Context, header RequestHeader, body io.Reader) error
	ResponseHeader(ctx context.Context) (ResponseHeader, error)
	ResponseBody(ctx context.Context) (io.Reader, error)

	Close() error
}

type ServerTransport interface {
	RequestHeader(ctx context.Context) (RequestHeader, error)
	RequestBody(ctx context.Context) (io.Reader, error)
	WriteResponse(ctx context.Context, header ResponseHeader, body io.Reader) error

	// Close can be called multiple times and must be idempotent.
	Close() error
}
