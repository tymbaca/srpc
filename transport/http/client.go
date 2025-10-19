package httptransport

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"net/url"

	"github.com/tymbaca/srpc"
	"github.com/tymbaca/srpc/pkg/sps"
)

func NewClientConnector(path string, method string) srpc.Connector {
	return &Connector{
		path:   path,
		method: method,
		client: &http.Client{},
		spsKey: nil,
	}
}

func NewClientConnectorSPS(path string, method string, privateKey string) (srpc.Connector, error) {
	privateKeyRaw, err := hex.DecodeString(privateKey)
	if err != nil {
		return nil, fmt.Errorf("privateKey must be hex: %w", err)
	}

	spsKey := new(big.Int).SetBytes(privateKeyRaw)

	return &Connector{
		path:   path,
		method: method,
		client: &http.Client{},
		spsKey: spsKey,
	}, nil
}

type Connector struct {
	path   string
	method string
	client *http.Client

	spsKey *big.Int
}

func (cl *Connector) Connect(ctx context.Context, addr string) (srpc.ClientConn, error) {
	url, err := url.JoinPath(addr, cl.path)
	if err != nil {
		return nil, fmt.Errorf("create url to connect via http: %w", err)
	}

	return &clientConn{
		url:    url,
		method: cl.method,
		client: cl.client,
	}, nil
}

type clientConn struct {
	url    string
	method string
	client *http.Client
	sps    *sps.SPS
	close  func() error
}

func (cl *clientConn) Do(ctx context.Context, req srpc.Request) (srpc.Response, error) {
	httpReq, err := http.NewRequestWithContext(ctx, cl.method, cl.url, req.Body)
	if err != nil {
		return srpc.Response{}, fmt.Errorf("create http request: %w", err)
	}

	// WARN: is it safe to set a whole header like this?
	httpReq.Header, err = toHeader(req.ServiceMethod, req.Metadata)
	if err != nil {
		return srpc.Response{}, fmt.Errorf("encode req header: %w", err)
	}

	httpResp, err := cl.client.Do(httpReq)
	if err != nil {
		return srpc.Response{}, fmt.Errorf("do http request: %w", err)
	}

	var resp srpc.Response

	if httpResp.StatusCode != http.StatusOK {
		respBody, err := io.ReadAll(httpResp.Body)
		if err != nil {
			return srpc.Response{}, fmt.Errorf("got bad status code: %s, cannot ready response body: %w", httpResp.Status, err)
		}
		resp.Error = fmt.Errorf("got bad status code: %s, body: %s", httpResp.Status, respBody)
		return resp, nil
	}

	if httpResp.Body != nil {
		cl.close = httpResp.Body.Close
	}

	resp.StatusCode, err = getStatus(httpResp.Header)
	if err != nil {
		return srpc.Response{}, fmt.Errorf("get status from resp header: %w", err)
	}

	resp.ServiceMethod, resp.Metadata, err = fromHeader(httpResp.Header)
	if err != nil {
		return srpc.Response{}, fmt.Errorf("decode resp header: %w", err)
	}

	if hasError(httpResp.Header) {
		errMsg, err := io.ReadAll(httpResp.Body)
		if err != nil {
			return srpc.Response{}, fmt.Errorf("read error from response: %w", err)
		}

		resp.Error = errors.New(string(errMsg))
	} else {
		resp.Body = httpResp.Body
	}

	return resp, nil
}

// Close must be called after Send
func (cl *clientConn) Close() error {
	return cl.close()
}
