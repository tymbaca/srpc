package tcp

import (
	"context"
	"net"

	"github.com/tymbaca/srpc"
)

type Connector struct{}

func (co *Connector) Connect(ctx context.Context, addr string) (srpc.Conn, error) {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, err
	}

	return &Conn{c: conn}, nil
}

type Conn struct {
	c net.Conn
}

func (c *Conn) RemoteAddr() string {
	return c.c.RemoteAddr().String()
}

func (c *Conn) LocalAddr() string {
	return c.c.LocalAddr().String()
}

func (c *Conn) Read(p []byte) (n int, err error) {
	return c.c.Read(p)
}

func (c *Conn) Write(p []byte) (n int, err error) {
	return c.c.Write(p)
}

func (c *Conn) Close() error {
	return c.c.Close()
}
