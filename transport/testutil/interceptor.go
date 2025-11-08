package testutil

import (
	"bytes"
	"context"

	"github.com/tymbaca/srpc"
)

func NewDialInterceptor(dialer srpc.Dialer) (*InterceptorDialer, *InterceptorConn) {
	interceptor := &InterceptorConn{
		WData: &bytes.Buffer{},
		RData: &bytes.Buffer{},
	}

	interceptorDialer := &InterceptorDialer{
		interceptor: interceptor,
		Dialer:      dialer,
	}

	return interceptorDialer, interceptor
}

type InterceptorDialer struct {
	interceptor *InterceptorConn
	srpc.Dialer
}

func (d *InterceptorDialer) Dial(ctx context.Context, addr string) (srpc.Conn, error) {
	conn, err := d.Dialer.Dial(ctx, addr)
	if err != nil {
		return nil, err
	}

	d.interceptor.Conn = conn
	return d.interceptor, nil
}

type InterceptorConn struct {
	WData *bytes.Buffer
	RData *bytes.Buffer
	srpc.Conn
}

func (c *InterceptorConn) Read(p []byte) (n int, err error) {
	n, err = c.Conn.Read(p)
	c.RData.Write(p[:n])
	return n, err
}

func (c *InterceptorConn) Write(p []byte) (n int, err error) {
	n, err = c.Conn.Write(p)
	c.WData.Write(p[:n])
	return n, err
}
