package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/tymbaca/srpc"
	"github.com/tymbaca/srpc/codec/json"
	"github.com/tymbaca/srpc/logger"
	"github.com/tymbaca/srpc/transport"
	"github.com/tymbaca/srpc/transport/stdnet"
)

//go:generate srpc-gen --target=EchoService
type EchoService interface {
	Echo(ctx context.Context, req string) (string, error)
	EchoStruct(ctx context.Context, req EchoStructReq) (EchoStructResp, error)
}

type EchoStructReq struct {
	Val string
}

type EchoStructResp struct {
	Val string
}

func main() {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	addr := "localhost:8080"

	go func() {
		server := NewEchoServiceServer(srpc.NewServer(
			json.Codec,
			srpc.WithStreamingResponse(true),
			srpc.WithLogger(logger.DefaultSLogger{}),
		))
		defer server.Close()

		l, err := stdnet.Listen("tcp", addr)
		if err != nil {
			panic(err)
		}

		err = server.Start(ctx, l)
		if errors.Is(err, transport.ErrListenerClosed) {
			return
		} else if err != nil {
			slog.Error(err.Error())
		}
	}()

	client := NewEchoServiceClient(srpc.NewClient(addr, json.Codec, stdnet.NewDialer("tcp")))

	fmt.Println("Echo")
	resp, err := client.Echo(ctx, "hello")
	if err != nil {
		panic(err)
	}

	fmt.Println("msg from server:", resp)

	fmt.Println("EchoStruct")
	respStruct, err := client.EchoStruct(ctx, EchoStructReq{"hello"})
	if err != nil {
		panic(err)
	}

	fmt.Println("msg from server:", respStruct)
}
