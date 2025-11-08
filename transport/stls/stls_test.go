package stls

import (
	"bytes"
	"context"
	"io"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tymbaca/srpc"
	"github.com/tymbaca/srpc/transport/inmem"
)

func TestInside(t *testing.T) {
	ctx := t.Context()

	cluster := inmem.New()
	clientPeer := cluster.NewPeer()
	serverPeer := cluster.NewPeer()

	clientInterceptor := &interceptorConn{
		wdata: &bytes.Buffer{},
		rdata: &bytes.Buffer{},
	}

	interceptorDialer := &interceptorDialer{
		interceptor: clientInterceptor,
		Dialer:      clientPeer,
	}

	dialer, err := NewDialerRandomKey(interceptorDialer)
	require.NoError(t, err)

	listener, err := NewListenerRandomKey(serverPeer.Listen())
	require.NoError(t, err)

	clientMsg := bytes.Repeat([]byte("c"), 100)
	serverMsg := bytes.Repeat([]byte("s"), 100)

	go func() {
		conn, err := listener.Accept()
		require.NoError(t, err)
		defer conn.Close()

		buf := make([]byte, len(clientMsg))
		_, err = io.ReadFull(conn, buf)
		require.NoError(t, err)
		require.Equal(t, clientMsg, buf)

		_, err = conn.Write(serverMsg)
		require.NoError(t, err)
	}()

	conn, err := dialer.Dial(ctx, serverPeer.Addr())
	require.NoError(t, err)

	_, err = conn.Write(clientMsg)
	require.NoError(t, err)

	buf := make([]byte, len(serverMsg))
	_, err = io.ReadFull(conn, buf)
	require.NoError(t, err)
	require.Equal(t, serverMsg, buf)

	{
		handshakeLen := 1 + 1 + 32

		wdata := clientInterceptor.wdata.Bytes()
		require.Len(t, wdata, handshakeLen+len(clientMsg))
		require.NotEqual(t, clientMsg, wdata[handshakeLen:])

		rdata := clientInterceptor.rdata.Bytes()
		require.Len(t, rdata, handshakeLen+len(serverMsg))
		require.NotEqual(t, serverMsg, rdata[handshakeLen:])
	}
}

type interceptorDialer struct {
	interceptor *interceptorConn
	srpc.Dialer
}

func (d *interceptorDialer) Dial(ctx context.Context, addr string) (srpc.Conn, error) {
	conn, err := d.Dialer.Dial(ctx, addr)
	if err != nil {
		return nil, err
	}

	d.interceptor.Conn = conn
	return d.interceptor, nil
}

type interceptorConn struct {
	wdata *bytes.Buffer
	rdata *bytes.Buffer
	srpc.Conn
}

func (c *interceptorConn) Read(p []byte) (n int, err error) {
	n, err = c.Conn.Read(p)
	c.rdata.Write(p[:n])
	return n, err
}

func (c *interceptorConn) Write(p []byte) (n int, err error) {
	n, err = c.Conn.Write(p)
	c.wdata.Write(p[:n])
	return n, err
}
