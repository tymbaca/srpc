package srpc

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/tymbaca/srpc/pkg/pipe"
)

var (
	ErrServiceError   = errors.New("service error")
	ErrTransportError = errors.New("transport error")
)

func NewClient(addr string, codec Codec, connector Connector) *Client {
	return &Client{
		addr:      addr,
		codec:     codec,
		connector: connector,
	}
}

type Client struct {
	addr      string
	codec     Codec
	connector Connector
}

// TODO: check metadata in context
// TODO: timeouts? > but we support context

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

	if connResp.StatusCode != StatusOK {
		coreErr := ErrTransportError
		if connResp.StatusCode == StatusErrorFromService {
			coreErr = ErrServiceError
		}
		if connResp.Error != nil {
			return fmt.Errorf("%w: %s", coreErr, connResp.Error)
		} else {
			return fmt.Errorf("%w: (no error message)", coreErr)
		}
	}

	err = c.codec.Decode(connResp.Body, resp)
	if err != nil {
		return fmt.Errorf("decode response body: %w", err)
	}

	return nil
}
