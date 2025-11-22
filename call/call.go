package call

import (
	"context"

	"github.com/tymbaca/srpc/internal/callsuite"
	"github.com/tymbaca/srpc/metadata"
)

// Option modifies current call.
type Option func(c *callsuite.Suite)

type callOptKey struct{}

// OptionsFromContext retreives call options from context. It's for internal use.
func OptionsFromContext(ctx context.Context) []Option {
	opts, _ := ctx.Value(callOptKey{}).([]Option)
	return opts
}

// WithOptions puts provided options into the context. [srpc.Client.Call] then
// retreives them and applies to the call.
func WithOptions(ctx context.Context, opts ...Option) context.Context {
	return context.WithValue(ctx, callOptKey{}, opts)
}

// WithResponseMetadata call option specifies pointer, where the response metadata will be stored.
func WithResponseMetadata(md *metadata.Metadata) Option {
	return func(c *callsuite.Suite) {
		c.RespMetadata = md
	}
}
