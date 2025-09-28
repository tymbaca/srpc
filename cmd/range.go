package main

import (
	"context"
	"encoding/json"
	"io"

	"github.com/tymbaca/srpc"
	"github.com/tymbaca/srpc/codechelp"
	"github.com/tymbaca/srpc/pkg/pipe"
	"github.com/tymbaca/srpc/transport/httptransport"
)

func fn() {
	ctx := context.Background()
	jsonEnc := codechelp.ToEncoder(json.NewEncoder)
	httpClientTransport := httptransport.NewClientTransport()

	type Node struct {
		ID string
	}

	node := Node{ID: "node1"}
	reqHeader := srpc.RequestHeader{
		ServiceMethod: "Service.Method",
		Metadata:      map[string]string{"hello": "world"},
	}

	err := httpClientTransport.WriteRequest(ctx,
		reqHeader,
		pipe.ToReader(func(w io.Writer) error {
			return jsonEnc.Encode(w, node)
		}),
	)
	if err != nil {
		panic(err)
	}

	////////////////////

	jsonDec := codechelp.ToDecoder(json.NewDecoder)
}
