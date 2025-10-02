package testpackage

import "context"

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

//go:generate srpc-gen --target=TestService
type TestService interface {
	Add(ctx context.Context, req AddReq) (AddResp, error)
	Divide(ctx context.Context, req DivideReq) (DivideResp, error)
}
