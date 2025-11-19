package srpc

import (
	"context"
	"fmt"
	"io"

	"github.com/tymbaca/srpc/call"
	"github.com/tymbaca/srpc/codec"
	"github.com/tymbaca/srpc/enc"
	"github.com/tymbaca/srpc/internal/callsuite"
	"github.com/tymbaca/srpc/metadata"
	"github.com/tymbaca/srpc/pkg/fx"
	"github.com/tymbaca/srpc/pkg/pipe"
	"github.com/tymbaca/srpc/status"
	"github.com/tymbaca/srpc/transport"
)

var encVersion = enc.Version{Major: 0, Minor: 1, Patch: 0}

// NewClient create a new [Client] with provided address, codec and connector.
func NewClient(addr string, codec codec.Codec, connector transport.Dialer) *Client {
	return &Client{
		addr:      addr,
		enc:       enc.Context{Version: encVersion, IgnoreVersion: false},
		codec:     codec,
		connector: connector,
	}
}

// Client is used to invoke RPC's on the server.
type Client struct {
	addr      string
	enc       enc.Context
	codec     codec.Codec
	connector transport.Dialer
}

// Call invokes the serviceMethod RPC on the server. Provided req and resp must
// match the input and output type on the server, resp must be a valid pointer.
// They will be encoded/decoded using the specified client codec.
func (c *Client) Call(ctx context.Context, serviceMethod string, req any, resp any) error {
	conn, err := c.connector.Dial(ctx, c.addr)
	if err != nil {
		return fmt.Errorf("connect %s: %w", c.addr, err)
	}
	_, cancel := fx.CloseOnCancel(ctx, conn)
	defer cancel()

	md, _ := metadata.FromContext(ctx)
	encReq := enc.Request{
		ServiceMethod: enc.NewString(serviceMethod),
		Metadata:      enc.NewMetadata(md),
		Body: pipe.ToReader(func(w io.Writer) error {
			return c.codec.Encode(w, req)
		}),
	}

	callSuite := callsuite.Suite{
		Req:          encReq,
		Conn:         conn,
		RespMetadata: nil,
	}

	for _, opt := range call.OptionsFromContext(ctx) {
		opt(&callSuite)
	}

	err = enc.WriteRequest(c.enc, callSuite.Conn, callSuite.Req)
	if err != nil {
		return err
	}

	connResp, err := enc.ReadResponse(c.enc, conn)
	if err != nil {
		return err
	}
	defer drain(connResp.Body)

	if callSuite.RespMetadata != nil {
		*callSuite.RespMetadata = connResp.Metadata.Map()
	}

	if connResp.StatusCode != status.OK {
		errMsg := connResp.Error.Error()
		if errMsg == "" {
			return status.ErrorOnlyCode(connResp.StatusCode)
		}
		return status.Error(connResp.StatusCode, errMsg)
	}

	err = c.codec.Decode(connResp.Body, resp)
	if err != nil {
		return status.Errorf(status.InternalError, "decode response body: %w", err)
	}

	return nil
}
