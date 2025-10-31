package srpc

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/tymbaca/srpc/pkg/enc"
	"github.com/tymbaca/srpc/pkg/pipe"
)

var (
	ErrServiceError   = errors.New("service error")
	ErrTransportError = errors.New("transport error")
)

var encVersion = enc.Version{Major: 0, Minor: 1, Patch: 0}

func NewClient(addr string, codec Codec, connector Connector) *Client {
	return &Client{
		addr:         addr,
		payloadcodec: codec,
		connector:    connector,
	}
}

type Client struct {
	addr         string
	headerCodec  Codec
	payloadcodec Codec
	connector    Connector
}

func (c *Client) encoder() enc.Codec {
	return enc.Codec{Version: encVersion, IgnoreVersion: false}
}

// TODO: check metadata in context
// TODO: timeouts? > but we support context

func (c *Client) Call(ctx context.Context, serviceMethod string, req any, resp any) error {
	conn, err := c.connector.Connect(ctx, c.addr)
	if err != nil {
		return fmt.Errorf("connect %s: %w", c.addr, err)
	}

	encReq := enc.Request{
		ServiceMethod: enc.NewString(serviceMethod),
		Metadata:      enc.Metadata{}, // TODO:
		Body:          pipe.ToReader(func(w io.Writer) error { return c.payloadcodec.Encode(w, req) }),
	}
	if err != nil {
		return fmt.Errorf("send request: %w", err)
	}

	e := c.encoder()

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

	err = c.payloadcodec.Decode(connResp.Body, resp)
	if err != nil {
		return fmt.Errorf("decode response body: %w", err)
	}

	return nil
}
