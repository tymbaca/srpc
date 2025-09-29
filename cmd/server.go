package main

import (
	"context"

	"github.com/tymbaca/srpc"
)

type (
	GetNodesReq struct {
		arg int
	}
	GetNodesResp struct {
		val string
	}
	CreateNodesReq  struct{}
	CreateNodesResp struct{}
)

type MyService interface {
	GetNodes(ctx context.Context, req GetNodesReq) (GetNodesResp, error)
	CreateNodes(ctx context.Context, req CreateNodesReq) (CreateNodesResp, error)
	// ...
}

type ServerImpl struct{}

func (s *ServerImpl) GetNodes(ctx context.Context, req GetNodesReq) (GetNodesResp, error) {
	panic("not implemented")
}

func (s *ServerImpl) CreateNodes(ctx context.Context, req CreateNodesReq) (CreateNodesResp, error) {
	panic("not implemented")
}

func genericserver() {
	impl := &ServerImpl{}
	srpc.NewServer[MyService](impl)
}
