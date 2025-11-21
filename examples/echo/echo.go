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
	Echo(ctx context.Context, input string) (string, error)
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

	resp, err := client.Echo(ctx, "hello")
	if err != nil {
		panic(err)
	}

	fmt.Println("msg from server:", resp)
}
