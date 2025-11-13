package testutil

import (
	"context"

	"github.com/tymbaca/srpc/transport"
)

type interruptConfig struct {
	CloseAfterDial bool
}

func newInterruptDialer(ctx context.Context, d transport.Dialer, cfg *interruptConfig) (wrapped transport.Dialer, cancel func()) {
	ctx, cancel = context.WithCancel(ctx)

	return &interruptDialer{
		backing: d,
		ctx:     ctx,
		cancel:  cancel,
		cfg:     cfg,
	}, cancel
}

type interruptDialer struct {
	backing transport.Dialer
	ctx     context.Context
	cancel  func()
	cfg     *interruptConfig
}

func (d *interruptDialer) Dial(ctx context.Context, addr string) (transport.Conn, error) {
	if err := d.ctx.Err(); err != nil {
		return nil, err
	}

	conn, err := d.backing.Dial(ctx, addr)
	if err != nil {
		return nil, err
	}

	if d.cfg != nil || d.cfg.CloseAfterDial {
		d.cancel()
	}

	return &interruptConn{
		backing: conn,
		ctx:     d.ctx,
	}, nil
}

type interruptConn struct {
	backing transport.Conn
	ctx     context.Context
}

func (c *interruptConn) RemoteAddr() string {
	if err := c.ctx.Err(); err != nil {
		return ""
	}

	return c.backing.RemoteAddr()
}

func (c *interruptConn) LocalAddr() string {
	if err := c.ctx.Err(); err != nil {
		return ""
	}

	return c.backing.LocalAddr()
}

func (c *interruptConn) Read(p []byte) (n int, err error) {
	if err := c.ctx.Err(); err != nil {
		return 0, err
	}

	return c.backing.Read(p)
}

func (c *interruptConn) Write(p []byte) (n int, err error) {
	if err := c.ctx.Err(); err != nil {
		return 0, err
	}

	return c.backing.Write(p)
}

func (c *interruptConn) Close() error {
	if err := c.ctx.Err(); err != nil {
		return err
	}

	return c.backing.Close()
}
