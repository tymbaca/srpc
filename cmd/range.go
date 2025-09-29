package main

// import (
// 	"context"
// 	"encoding/json"
// 	"io"
// 	"net/rpc"
//
// 	"github.com/tymbaca/srpc"
// 	"github.com/tymbaca/srpc/codechelp"
// 	"github.com/tymbaca/srpc/pkg/pipe"
// 	"github.com/tymbaca/srpc/transport/httptransport"
// )
//
// func client() {
// 	client := srpc.NewClient(addr)
// 	var req any
// 	var resp any
// 	client.Call(ctx, serviceMethod, req, &resp)
// }
//
// type (
// 	Method1Req  struct{}
// 	Method1Resp struct{}
// )
//
// func server() {
// 	l, err := httptransport.Listen(addr)
// 	if err != nil {
// 		panic(err)
// 	}
//
// 	server := srpc.NewServer(codechelp.T)
// 	server.Handle("Service.Method1", func(ctx context.Context, req Method1Req) (Method1Resp, error) {
// 	})
// 	server.Accept(l)
// }
//
// func rpctest() {
// 	rpc.Accept()
// }
//
// func fn() {
// 	ctx := context.Background()
// 	jsonEnc := codechelp.ToEncoder(json.NewEncoder)
// 	httpClientTransport := httptransport.NewClientTransport()
//
// 	type Node struct {
// 		ID string
// 	}
//
// 	node := Node{ID: "node1"}
// 	reqHeader := srpc.Request{
// 		ServiceMethod: "Service.Method",
// 		Metadata:      map[string]string{"hello": "world"},
// 	}
//
// 	err := httpClientTransport.WriteRequest(
// 		reqHeader,
// 		pipe.ToReader(func(w io.Writer) error {
// 			return jsonEnc.Encode(w, node)
// 		}),
// 	)
// 	if err != nil {
// 		panic(err)
// 	}
//
// 	////////////////////
//
// 	// jsonDec := codechelp.ToDecoder(json.NewDecoder)
// 	// httpClientTransport.
// }
