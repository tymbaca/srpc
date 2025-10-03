package main

import (
	"context"

	"github.com/tymbaca/srpc/cmd/tymbaca/inner"
)

// NOTE: srpc-gen - is a code generation tool that creates type-safe client and
// server implementation. You can check generated sibling files in this folder ("srpc.*" files).
// To generate the files use `go generate ./...` command.

//go:generate srpc-gen --target=TestService
type TestService interface {
	Add(ctx context.Context, req AddReq) (AddResp, error)
	Divide(ctx context.Context, req DivideReq) (DivideResp, error)
	// NOTE: input and output types can be from another packages
	Multiply(ctx context.Context, req inner.MultiplyReq) (inner.MultiplyResp, error)
}

type (
	AddReq struct {
		A, B int
	}
	AddResp struct {
		Result int
	}
)

type (
	DivideReq struct {
		A, B int
	}
	DivideResp struct {
		Result int
	}
)
