package main

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"log/slog"
	"sync"

	"github.com/tymbaca/srpc"
	"github.com/tymbaca/srpc/codec/json"
	"github.com/tymbaca/srpc/logger"
	"github.com/tymbaca/srpc/transport"
	"github.com/tymbaca/srpc/transport/inmem"
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
	mainTCP()
	mainInmem()
}

func mainTCP() {
	var wg sync.WaitGroup
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	addr := "localhost:8080"

	server := NewEchoServiceServer(srpc.NewServer(
		json.Codec,
		srpc.WithStreamingResponse(true),
		srpc.WithLogger(logger.DefaultSLogger{}),
	))

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

	wg.Add(1)
	go func() {
		defer wg.Done()

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
	d, err = stls.NewDialerRandomKey(d, rand.Reader)
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
	// happen, uncomment to try
	// {
	// 	client := NewEchoServiceClient(srpc.NewClient(addr, json.Codec, stdnet.NewDialer("tcp")))
	//
	// 	_, err := client.Echo(ctx, "hello")
	// 	if err != nil {
	// 		fmt.Println("client without sTLS, err:", err)
	// 	}
	// }

	server.Close()
	wg.Wait()
}

func mainInmem() {
	var wg sync.WaitGroup
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	cluster := inmem.New()
	serverPeer := cluster.NewPeer()
	clientPeer := cluster.NewPeer()

	server := NewEchoServiceServer(srpc.NewServer(
		json.Codec,
		srpc.WithStreamingResponse(true),
		srpc.WithLogger(logger.DefaultSLogger{}),
	))

	// Wrap listener with sTLS
	l, err := stls.NewListenerRandomKey(serverPeer.Listen(), rand.Reader)
	if err != nil {
		panic(err)
	}

	wg.Add(1)
	go func() {
		defer wg.Done()

		err = server.Start(ctx, l)
		if errors.Is(err, transport.ErrListenerClosed) {
			return
		} else if err != nil {
			slog.Error(err.Error())
		}
	}()

	// Wrap dialer with sTLS
	d, err := stls.NewDialerRandomKey(clientPeer, rand.Reader)
	if err != nil {
		panic(err)
	}

	client := NewEchoServiceClient(srpc.NewClient(serverPeer.Addr(), json.Codec, d))

	resp, err := client.Echo(ctx, "hello")
	if err != nil {
		panic(err)
	}

	fmt.Println("msg from server:", resp)

	// here we can try to create the client without sTLS and see what will
	// happen, uncomment to try
	// {
	// 	client := NewEchoServiceClient(srpc.NewClient(serverPeer.Addr(), json.Codec, clientPeer))
	//
	// 	_, err := client.Echo(ctx, "hello")
	// 	if err != nil {
	// 		fmt.Println("client without sTLS, err:", err)
	// 	}
	// }

	server.Close()
	wg.Wait()
}
