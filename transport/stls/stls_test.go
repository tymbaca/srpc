package stls

import (
	"bytes"
	"io"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tymbaca/srpc"
	"github.com/tymbaca/srpc/transport/inmem"
	"github.com/tymbaca/srpc/transport/testutil"
	"go.uber.org/goleak"
	"golang.org/x/crypto/chacha20"
)

func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m)
}

func newListener(c *inmem.Cluster) func() srpc.Listener {
	return func() srpc.Listener {
		l, err := NewListenerRandomKey(c.NewPeer().Listen())
		if err != nil {
			panic(err)
		}
		return l
	}
}

func newDialer(c *inmem.Cluster) func() srpc.Dialer {
	return func() srpc.Dialer {
		d, err := NewDialerRandomKey(c.NewPeer())
		if err != nil {
			panic(err)
		}
		return d
	}
}

func TestSimple(t *testing.T) {
	c := inmem.New()
	testutil.TestSimple(t,
		newListener(c),
		newDialer(c),
	)
}

func TestStress(t *testing.T) {
	c := inmem.New()
	testutil.TestStress(t,
		newListener(c),
		newDialer(c),
		100,
		100,
	)
}

func BenchmarkStress(b *testing.B) {
	c := inmem.New()
	testutil.Benchmark(b,
		newListener(c),
		newDialer(c),
	)
}

func TestInside(t *testing.T) {
	ctx := t.Context()

	cluster := inmem.New()
	clientPeer := cluster.NewPeer()
	serverPeer := cluster.NewPeer()

	interceptorDialer, interceptor := testutil.NewDialInterceptor(clientPeer)

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
		pubKeyLen := 65
		handshakeLen := 1 + 1 + pubKeyLen + 1 + chacha20.NonceSizeX

		wdata := interceptor.WData.Bytes()
		require.Len(t, wdata, handshakeLen+len(clientMsg))
		require.NotEqual(t, clientMsg, wdata[handshakeLen:])

		rdata := interceptor.RData.Bytes()
		require.Len(t, rdata, handshakeLen+len(serverMsg))
		require.NotEqual(t, serverMsg, rdata[handshakeLen:])
	}
}
