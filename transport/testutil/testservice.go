package testutil

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

type (
	ReplyMDReq struct {
		Key string
	}
	ReplyMDResp struct {
		Vals []string
		Ok   bool
	}
)

//go:generate srpc-gen --target=TestService
type TestService interface {
	Add(ctx context.Context, req AddReq) (AddResp, error)
	LongAdd(ctx context.Context, req AddReq) (AddResp, error)
	Divide(ctx context.Context, req DivideReq) (DivideResp, error)
	ReplyMD(ctx context.Context, req ReplyMDReq) (ReplyMDResp, error)
}
