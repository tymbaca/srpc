package testpackage

import (
	"context"
	"fmt"
	"net/http"

	"github.com/tymbaca/srpc"
	"github.com/tymbaca/srpc/cmd/testpackage/inner"
	"github.com/tymbaca/srpc/codec"
	httptransport "github.com/tymbaca/srpc/transport/http"
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

//go:generate srpc-gen --target=TestService
type TestService interface {
	Add(ctx context.Context, req AddReq) (AddResp, error)
	Divide(ctx context.Context, req DivideReq) (DivideResp, error)
}

//go:generate srpc-gen --target=TestService2
type TestService2 interface {
	Divide(ctx context.Context, req DivideReq) (DivideResp, error)
	Multiply(ctx context.Context, req inner.MultiplyReq) (inner.MultiplyResp, error)
}

func exampleServer() {
	ctx := context.Background()
	err := NewTestServiceServer(codec.JSON).Start(ctx, httptransport.CreateAndStartListener(":8080", "/srpc", http.MethodPost))
	if err != nil {
		panic(err)
	}
}

func exampleClient() {
	ctx := context.Background()
	client := NewTestServiceClient(srpc.NewClient("localhost:8080", codec.JSON, httptransport.NewClientConnector("/srpc", http.MethodPost)))

	resp, err := client.Divide(ctx, DivideReq{A: 10, B: 2})
	if err != nil {
		panic(err)
	}

	fmt.Printf("resp: %v\n", resp.Result) // 5
}
