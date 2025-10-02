package main

import (
	"context"
	"fmt"
)

type (
	GetNodesReq struct {
		Arg int
	}
	GetNodesResp struct {
		Val string
	}
	CreateNodesReq  struct{}
	CreateNodesResp struct{}
)

type MyService interface {
	GetNodes(ctx context.Context, req GetNodesReq) (GetNodesResp, error)
	CreateNodes(ctx context.Context, req CreateNodesReq) (CreateNodesResp, error)
	// ...
}

type ServerImpl struct {
	A int
}

func (s *ServerImpl) GetNodes(ctx context.Context, req GetNodesReq) (GetNodesResp, error) {
	// return GetNodesResp{strconv.Itoa(req.arg)}, nil
	return GetNodesResp{}, fmt.Errorf("some err")
}

func (s *ServerImpl) CreateNodes(ctx context.Context, req CreateNodesReq) (CreateNodesResp, error) {
	panic("not implemented")
}

// func genericserver() {
// 	impl := &ServerImpl{}
// 	srpc.NewServer[MyService](impl)
// }
