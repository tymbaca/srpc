package main

import (
	"context"
	"fmt"

	"github.com/tymbaca/srpc"
	"github.com/tymbaca/srpc/codec/json"
	"github.com/tymbaca/srpc/transport/inmem"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cluster := inmem.New()
	client := cluster.NewPeer()
	server := cluster.NewPeer()

	go runServer(ctx, server)
	runClient(ctx, client, server.Addr())
}

func runServer(ctx context.Context, peer *inmem.Peer) {
	server := NewTestServiceServer(srpc.NewServer(json.Codec))
	defer server.Close()

	err := server.Start(ctx, peer.Listen())
	if err != nil {
		panic(err)
	}
}

func runClient(ctx context.Context, peer *inmem.Peer, target string) {
	client := NewTestServiceClient(srpc.NewClient(target, json.Codec, peer))

	// insead of `client.Call(ctx, "TestService.Divide", req, &resp)`
	resp, err := client.Divide(ctx, DivideReq{A: 10, B: 2})
	if err != nil {
		panic(err)
	}

	fmt.Printf("resp: %v\n", resp.Result) // 5
}
