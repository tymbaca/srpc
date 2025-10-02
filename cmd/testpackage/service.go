package testpackage

import (
	"context"

	"github.com/tymbaca/srpc/cmd/testpackage/inner"
)

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

type TestService interface {
	Add(ctx context.Context, req AddReq) (AddResp, error)
	Divide(ctx context.Context, req DivideReq) (DivideResp, error)
}

//go:generate srpc-gen --target=TestService2
type TestService2 interface {
	Divide(ctx context.Context, req DivideReq) (DivideResp, error)
	Multiply(ctx context.Context, req inner.MultiplyReq) (inner.MultiplyResp, error)
}
