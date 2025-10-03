package main

import (
	"context"
	"fmt"
	"net/http"

	"github.com/tymbaca/srpc"
	"github.com/tymbaca/srpc/codec"
	httptransport "github.com/tymbaca/srpc/transport/http"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go runServer(ctx)
	runClient(ctx)
}

func runServer(ctx context.Context) {
	err := NewTestServiceServer(srpc.NewServer(codec.JSON)).Start(ctx, httptransport.CreateAndStartListener(":8080", "/srpc", http.MethodPost))
	if err != nil {
		panic(err)
	}
}

func runClient(ctx context.Context) {
	client := NewTestServiceClient(srpc.NewClient("localhost:8080", codec.JSON, httptransport.NewClientConnector("/srpc", http.MethodPost)))

	// insead of `client.Call(ctx, "TestService.Divide", req, &resp)`
	resp, err := client.Divide(ctx, DivideReq{A: 10, B: 2})
	if err != nil {
		panic(err)
	}

	fmt.Printf("resp: %v\n", resp.Result) // 5
}
