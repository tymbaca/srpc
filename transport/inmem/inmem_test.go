package inmem

import (
	"math/rand/v2"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/tymbaca/srpc"
	"github.com/tymbaca/srpc/codec"
	"github.com/tymbaca/srpc/logger"
	"github.com/tymbaca/srpc/transport/testdata"
	"go.uber.org/goleak"
)

func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m)
}

func TestInmemTransport(t *testing.T) {
	ctx := t.Context()

	cluster := New()
	clientPeer := cluster.NewPeer()
	serverPeer := cluster.NewPeer()

	server := testdata.NewTestServiceServer(srpc.NewServer(codec.JSON))
	defer server.Close()
	go server.Start(ctx, serverPeer.Listen())

	client := testdata.NewTestServiceClient(srpc.NewClient(serverPeer.Addr(), codec.JSON, clientPeer))
	{
		resp, err := client.Add(ctx, testdata.AddReq{A: 10, B: 15})
		require.NoError(t, err)
		require.Equal(t, 25, resp.Result)
	}
	{
		resp, err := client.Divide(ctx, testdata.DivideReq{A: 10, B: 2})
		require.NoError(t, err)
		require.Equal(t, 5, resp.Result)
	}
	{
		_, err := client.Divide(ctx, testdata.DivideReq{A: 10, B: 0})
		require.Error(t, err)
	}
}

func TestInmemTransportStress(t *testing.T) {
	ctx := t.Context()
	defer goleak.VerifyNone(t)

	t.Run("single client", func(t *testing.T) {
		cluster := New()
		serverPeer := cluster.NewPeer()
		server := testdata.NewTestServiceServer(srpc.NewServer(codec.JSON, srpc.WithLogger(logger.DefaulSLogger{})))
		defer server.Close()
		go server.Start(ctx, serverPeer.Listen())

		client := testdata.NewTestServiceClient(srpc.NewClient(serverPeer.Addr(), codec.JSON, cluster.NewPeer()))
		resp, err := client.Add(ctx, testdata.AddReq{A: 10, B: 15})
		require.NoError(t, err)
		require.Equal(t, 25, resp.Result)
	})

	t.Run("multiple clients parallel each multiple calls", func(t *testing.T) {
		cluster := New()
		serverPeer := cluster.NewPeer()
		server := testdata.NewTestServiceServer(srpc.NewServer(codec.JSON, srpc.WithLogger(logger.DefaulSLogger{})))
		defer server.Close()
		go server.Start(ctx, serverPeer.Listen())

		var wg sync.WaitGroup
		for range 100 {
			wg.Add(1)
			go func() {
				defer wg.Done()
				client := testdata.NewTestServiceClient(srpc.NewClient(serverPeer.Addr(), codec.JSON, cluster.NewPeer()))
				for range 100 {
					req := testdata.AddReq{A: rand.Int(), B: rand.Int()}
					resp, err := client.Add(ctx, req)
					require.NoError(t, err)
					require.Equal(t, req.A+req.B, resp.Result)
				}
			}()
		}
		wg.Wait()
	})

	t.Run("multiple clients parallel each multiple calls", func(t *testing.T) {
		cluster := New()
		serverPeer := cluster.NewPeer()
		server := testdata.NewTestServiceServer(srpc.NewServer(codec.JSON, srpc.WithLogger(logger.DefaulSLogger{})))
		defer server.Close()
		go server.Start(ctx, serverPeer.Listen())

		for range 10 {
			go func() {
				client := testdata.NewTestServiceClient(srpc.NewClient(serverPeer.Addr(), codec.JSON, cluster.NewPeer()))
				for {
					req := testdata.AddReq{A: rand.Int(), B: rand.Int()}
					_, err := client.Add(ctx, req)
					if err == nil {
						break
					}
				}
			}()
		}

		time.Sleep(50 * time.Millisecond)
		server.Close()
		time.Sleep(50 * time.Millisecond)
	})
}

func BenchmarkInmemTransportStress(b *testing.B) {
	ctx := b.Context()
	cluster := New()
	clientPeer := cluster.NewPeer()
	serverPeer := cluster.NewPeer()

	server := testdata.NewTestServiceServer(srpc.NewServer(codec.JSON, srpc.WithLogger(logger.DefaulSLogger{})))
	defer server.Close()
	go server.Start(ctx, serverPeer.Listen())

	client := testdata.NewTestServiceClient(srpc.NewClient(serverPeer.Addr(), codec.JSON, clientPeer))

	for b.Loop() {
		req := testdata.AddReq{A: rand.Int(), B: rand.Int()}
		resp, err := client.Add(ctx, req)
		_, _ = resp, err
	}
}
