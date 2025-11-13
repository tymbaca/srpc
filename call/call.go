package call

import (
	"context"

	"github.com/tymbaca/srpc/internal/callsuite"
	"github.com/tymbaca/srpc/metadata"
)

type Option func(c *callsuite.Suite)

type callOptKey struct{}

func WithOptions(ctx context.Context, opts ...Option) context.Context {
	return context.WithValue(ctx, callOptKey{}, opts)
}

func OptionsFromContext(ctx context.Context) []Option {
	opts, _ := ctx.Value(callOptKey{}).([]Option)
	return opts
}

func WithResponseMetadata(md *metadata.Metadata) Option {
	return func(c *callsuite.Suite) {
		c.RespMetadata = md
	}
}
