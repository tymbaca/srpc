package httptransport

import (
	"context"
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

type TestService interface {
	Add(ctx context.Context, req AddReq) (AddResp, error)
	// ...
}

type TestServiceImpl struct{}

func (s *TestServiceImpl) Add(ctx context.Context, req AddReq) (AddResp, error) {
	return AddResp{req.A + req.B}, nil
}

func TestHttpTransport(t *testing.T) {
	ctx := t.Context()
	svc := &TestServiceImpl{}

	server := srpc.NewServer(codec.JSON)
	srpc.Register[TestService](server, svc)
	go server.Start(ctx, CreateAndStartListener(":8080", "/srpc", http.MethodPost))

	client := srpc.NewClient("http://localhost:8080", codec.JSON, NewClientConnector("/srpc", http.MethodPost))

	var resp AddResp
	err := client.Call(ctx, "TestService.Add", AddReq{A: 10, B: 15}, &resp)
	require.NoError(t, err)
}
