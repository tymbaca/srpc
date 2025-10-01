package httptransport

import (
	"context"
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
}

func (cl *ClientConn) Send(ctx context.Context, req srpc.Request) (srpc.Response, error) {
	httpReq, err := http.NewRequestWithContext(ctx, cl.method, cl.url, req.Body)
	if err != nil {
		return srpc.Response{}, fmt.Errorf("create http request: %w", err)
	}

	httpReq.Header = http.Header(req.Metadata)

	httpResp, err := cl.client.Do(httpReq)
	if err != nil {
		return srpc.Response{}, fmt.Errorf("do http request: %w", err)
	}

	resp := srpc.Response{
		ServiceMethod: req.ServiceMethod,
		Metadata:      srpc.Metadata(httpResp.Header),
	}

	if httpResp.StatusCode != http.StatusOK {
		respBody, err := io.ReadAll(httpResp.Body)
		if err != nil {
			return srpc.Response{}, fmt.Errorf("cannot ready response body (status: %s): %w", httpResp.Status, err)
		}
		resp.Error = fmt.Errorf("got bad status code: %s, body: %s", httpResp.Status, respBody)
		return resp, nil
	}

	resp.Body = httpResp.Body
	return resp, nil
}
