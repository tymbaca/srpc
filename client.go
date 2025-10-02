package srpc

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/tymbaca/srpc/pkg/pipe"
)

var ErrServiceError = errors.New("service error")

func NewClient(addr string, connector ClientConnector, codec Codec) *Client {
	return &Client{
		addr:      addr,
		connector: connector,
		codec:     codec,
	}
}

type Client struct {
	addr      string
	connector ClientConnector
	codec     Codec
}

// TODO: check metadata in context
func (c *Client) Call(ctx context.Context, serviceMethod ServiceMethod, req any, resp any) error {
	conn, err := c.connector.Connect(ctx, c.addr)
	if err != nil {
		return fmt.Errorf("connect %s: %w", c.addr, err)
	}

	connResp, err := conn.Send(ctx, Request{
		ServiceMethod: serviceMethod,
		Metadata:      Metadata{}, // TODO:
		Body:          pipe.ToReader(func(w io.Writer) error { return c.codec.Encode(w, req) }),
	})
	if err != nil {
		return fmt.Errorf("send request: %w", err)
	}

	if connResp.Error != nil {
		return fmt.Errorf("%w: %s", ErrServiceError, connResp.Error)
	}

	err = c.codec.Decode(connResp.Body, resp)
	if err != nil {
		return fmt.Errorf("decode response body: %w", err)
	}

	return nil
}
