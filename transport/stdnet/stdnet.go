package stdnet

import (
	"context"
	"errors"
	"net"

	"github.com/tymbaca/srpc"
)

// Listen creates new [Listener] with provided network and address.
func Listen(network string, addr string) (*Listener, error) {
	l, err := net.Listen(network, addr)
	if err != nil {
		return nil, err
	}

	return &Listener{l: l}, nil
}

// Listener implements [net.Listener].
type Listener struct {
	l net.Listener
}

// Addr returns listener's address. Address is valid to use in [Dialer.Dial].
func (l *Listener) Addr() string {
	return l.l.Addr().String()
}

// Accept waits and returns new connection to the listener.
// If Listener got closed Accept must return [ErrListenerClosed],
// including Accept calls that didn't returned yet.
func (l *Listener) Accept() (srpc.Conn, error) {
	conn, err := l.l.Accept()
	if errors.Is(err, net.ErrClosed) {
		return nil, srpc.ErrListenerClosed
	}
	if err != nil {
		return nil, err
	}

	return &Conn{c: conn}, nil
}

// Close closes the listener.
// Any blocked Accept operations will be unblocked and return errors.
// Close can be called multiple times.
func (l *Listener) Close() error {
	return l.l.Close()
}

// NewDialer creates new [Dialer] with provided network.
func NewDialer(network string) *Dialer {
	return &Dialer{network: network}
}

// Dialer implements [srpc.Dialer].
type Dialer struct {
	network string
}

// Dial dials to addr using d.network.
func (d *Dialer) Dial(ctx context.Context, addr string) (srpc.Conn, error) {
	var dd net.Dialer
	conn, err := dd.DialContext(ctx, d.network, addr)
	if err != nil {
		return nil, err
	}

	return &Conn{c: conn}, nil
}

// Conn implements [srpc.Conn]
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
