package srpc

import (
	"context"

	"github.com/tymbaca/srpc/metadata"
	"github.com/tymbaca/srpc/pkg/enc"
)

type callSuite struct {
	ctx          context.Context
	req          enc.Request
	conn         Conn
	respMetadata *metadata.Metadata
}

type CallOption func(c *callSuite)

type callOptKey struct{}

func WithCallOptions(ctx context.Context, opts ...CallOption) context.Context {
	return context.WithValue(ctx, callOptKey{}, opts)
}

func getCallOptions(ctx context.Context) []CallOption {
	opts, _ := ctx.Value(callOptKey{}).([]CallOption)
	return opts
}

func WithResponseMetadata(md *metadata.Metadata) CallOption {
	return func(c *callSuite) {
		c.respMetadata = md
	}
}
