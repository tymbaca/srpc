// Package srpc provides simple and composable RPC implementation, similar to
// standard `net/rpc` but with `context.Context` support and type-safe
// client/server stub generation tool.
package srpc

import "github.com/tymbaca/srpc/enc"

// Version is current version of sRPC. Used in the core layer and can be used in
// transport implementations.
var Version = enc.Version{Major: 0, Minor: 1, Patch: 0}
