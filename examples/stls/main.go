package main

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"log/slog"

	"github.com/tymbaca/srpc"
	"github.com/tymbaca/srpc/codec/json"
	"github.com/tymbaca/srpc/logger"
	"github.com/tymbaca/srpc/transport"
	"github.com/tymbaca/srpc/transport/stdnet"
	"github.com/tymbaca/srpc/transport/stls"
)

//go:generate srpc-gen --target=EchoService
type EchoService interface {
	Echo(ctx context.Context, req string) (string, error)
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

		var l transport.Listener
		l, err := stdnet.Listen("tcp", addr)
		if err != nil {
			panic(err)
		}

		// Wrap listener with sTLS
		l, err = stls.NewListenerRandomKey(l, rand.Reader)
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

	var d transport.Dialer
	d = stdnet.NewDialer("tcp")

	// Wrap dialer with sTLS
	d, err := stls.NewDialerRandomKey(d, rand.Reader)
	if err != nil {
		panic(err)
	}

	client := NewEchoServiceClient(srpc.NewClient(addr, json.Codec, d))

	resp, err := client.Echo(ctx, "hello")
	if err != nil {
		panic(err)
	}

	fmt.Println("msg from server:", resp)

	// here we can try to create the client without sTLS and see what will
	// happen
	{
		client := NewEchoServiceClient(srpc.NewClient(addr, json.Codec, stdnet.NewDialer("tcp")))

		_, err := client.Echo(ctx, "hello")
		if err != nil {
			fmt.Println("client without sTLS, err:", err)
		}
	}
}
