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

func NewClient(addr string, codec Codec, connector Dialer) *Client {
	return &Client{
		addr:      addr,
		enc:       enc.Context{Version: encVersion, IgnoreVersion: false},
		codec:     codec,
		connector: connector,
	}
}

type Client struct {
	addr      string
	enc       enc.Context
	codec     Codec
	connector Dialer
}

// TODO: check metadata in context

func (c *Client) Call(ctx context.Context, serviceMethod string, req any, resp any) error {
	conn, err := c.connector.Dial(ctx, c.addr)
	if err != nil {
		return fmt.Errorf("connect %s: %w", c.addr, err)
	}
	_, cancel := closeOnCancel(ctx, conn)
	defer cancel()

	md, _ := MetadataFromContext(ctx)
	encReq := enc.Request{
		ServiceMethod: enc.NewString(serviceMethod),
		Metadata:      enc.NewMetadata(md),
		Body: pipe.ToReader(func(w io.Writer) error {
			return c.codec.Encode(w, req)
		}),
	}

	err = enc.WriteRequest(c.enc, conn, encReq)
	if err != nil {
		return err
	}

	connResp, err := enc.ReadResponse(c.enc, conn)
	if err != nil {
		return err
	}

	if connResp.StatusCode != enc.StatusOK {
		coreErr := ErrTransportError // TODO: remove. create function srpc.StatusFromError(err), so users can check status codes
		if connResp.StatusCode == enc.StatusErrorFromService {
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
