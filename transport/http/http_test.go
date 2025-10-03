package httptransport

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tymbaca/srpc"
	"github.com/tymbaca/srpc/codec"
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

type TestServiceImpl struct{}

func (s *TestServiceImpl) Add(ctx context.Context, req AddReq) (AddResp, error) {
	// return AddResp{}, errors.New("some err")
	return AddResp{req.A + req.B}, nil
}

func (s *TestServiceImpl) Divide(ctx context.Context, req DivideReq) (DivideResp, error) {
	if req.B == 0 {
		return DivideResp{}, errors.New("can't divide to 0")
	}

	return DivideResp{req.A / req.B}, nil
}

func TestHttpTransport(t *testing.T) {
	ctx := t.Context()
	svc := &TestServiceImpl{}

	server := srpc.NewServer(codec.JSON)
	srpc.Register[TestService](server, svc)
	// or:
	// srpc.RegisterWithName(server, svc, "TestService")
	go server.Start(ctx, CreateAndStartListener(":8080", "/srpc", http.MethodPost))

	client := srpc.NewClient("http://localhost:8080", codec.JSON, NewClientConnector("/srpc", http.MethodPost))
	{
		var resp AddResp
		err := client.Call(ctx, "TestService.Add", AddReq{A: 10, B: 15}, &resp)
		require.NoError(t, err)
		require.Equal(t, 25, resp.Result)
	}
	{
		var resp DivideResp
		err := client.Call(ctx, "TestService.Divide", DivideReq{A: 10, B: 2}, &resp)
		require.NoError(t, err)
		require.Equal(t, 5, resp.Result)
	}
	{
		var resp DivideResp
		err := client.Call(ctx, "TestService.Divide", DivideReq{A: 10, B: 0}, &resp)
		require.Error(t, err)
	}
}
