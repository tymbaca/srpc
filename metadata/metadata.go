// Package metadata provides sRPC Metadata type and helpers.
package metadata

import (
	"context"
)

// Metadata is user side metadata representation. It can be placed into the
// context in [Client.Call] calls, and extracted from context in server-side
// service implementation methods. It will be transfered by using
// [enc.Metadata] struture, see it for encoding details.
//
// Metadata can hold any strings, even invalid ones. They will be transfered as is.
//
// Keys starting with "srpc" are reserved for internal use.
type Metadata map[string][]string

type metadataCtxKey struct{}

func ToContext(ctx context.Context, md Metadata) context.Context {
	return context.WithValue(ctx, metadataCtxKey{}, md)
}

func FromContext(ctx context.Context) (Metadata, bool) {
	md, ok := ctx.Value(metadataCtxKey{}).(Metadata)
	return md, ok
}

type respMetadataCtxKey struct{}

// ResponseFromContext returns a pointer to outgoing response metadata. It can be used
// in server-side service implementation methods to fill outgoing metadata. If called
// from server-side service implementation methods the pointer would always be non-nil.
// Client can receive it by using [srpc.WithResponseMetadata] call option.
func ResponseFromContext(ctx context.Context) *Metadata {
	md, _ := ctx.Value(respMetadataCtxKey{}).(*Metadata)
	return md
}

// ResponseToContext is for internal use. It puts pointer to response data
// into the context. Then it can be retrieved by [ResponseFromContext].
func ResponseToContext(ctx context.Context, md *Metadata) context.Context {
	return context.WithValue(ctx, respMetadataCtxKey{}, md)
}
