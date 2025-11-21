package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/tymbaca/srpc"
	"github.com/tymbaca/srpc/call"
	"github.com/tymbaca/srpc/codec/json"
	"github.com/tymbaca/srpc/logger"
	"github.com/tymbaca/srpc/metadata"
	"github.com/tymbaca/srpc/transport"
	"github.com/tymbaca/srpc/transport/stdnet"
)

//go:generate srpc-gen --target=MetadataService
type MetadataService interface {
	EchoMetadata(ctx context.Context, req Empty) (Empty, error)
}

type Empty struct{}

func main() {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	addr := "localhost:8080"

	go func() {
		server := NewMetadataServiceServer(srpc.NewServer(
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

	client := NewMetadataServiceClient(srpc.NewClient(addr, json.Codec, stdnet.NewDialer("tcp")))

	// Put outgoing request metadata into the context
	ctx = metadata.ToContext(ctx, metadata.Metadata{
		"hello": {"world"},
	})

	// Specify option to store incoming response metadata
	var respMD metadata.Metadata
	ctx = call.WithOptions(ctx, call.WithResponseMetadata(&respMD))

	_, err := client.EchoMetadata(ctx, Empty{})
	if err != nil {
		panic(err)
	}

	fmt.Println("got md from server", respMD)
}
