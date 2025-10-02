package httptransport

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/tymbaca/srpc"
)

func NewClientConnector(path string, method string) srpc.ClientConnector {
	return &ClientConnector{
		path:   path,
		method: method,
		client: &http.Client{},
	}
}

type ClientConnector struct {
	path   string
	method string
	client *http.Client
}

func (cl *ClientConnector) Connect(ctx context.Context, addr string) (srpc.ClientConn, error) {
	url, err := url.JoinPath(addr, cl.path)
	if err != nil {
		return nil, fmt.Errorf("create url to connect via http: %w", err)
	}

	return &ClientConn{
		url:    url,
		method: cl.method,
		client: cl.client,
	}, nil
}

type ClientConn struct {
	url    string
	method string
	client *http.Client
	close  func() error
}

func (cl *ClientConn) Send(ctx context.Context, req srpc.Request) (srpc.Response, error) {
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

	resp.ServiceMethod, resp.Metadata, err = fromHeader(httpReq.Header)
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
func (cl *ClientConn) Close() error {
	return cl.close()
}
