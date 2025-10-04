package inmem

import (
	"sync"
	"testing"

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

func BenchmarkHttpTransport(b *testing.B) {
	ctx := b.Context()

	b.Run("single client", func(b *testing.B) {
		cluster := New()
		clientPeer := cluster.NewPeer()
		serverPeer := cluster.NewPeer()
		server := testdata.NewTestServiceServer(srpc.NewServer(codec.JSON, srpc.WithLogger(logger.DefaulSLogger{})))
		defer server.Close()
		go server.Start(ctx, serverPeer.Listen())

		client := testdata.NewTestServiceClient(srpc.NewClient(serverPeer.Addr(), codec.JSON, clientPeer))
		for b.Loop() {
			resp, err := client.Add(ctx, testdata.AddReq{A: 10, B: 15})
			require.NoError(b, err)
			require.Equal(b, 25, resp.Result)
		}
	})

	b.Run("multiple clients sequentual", func(b *testing.B) {
		cluster := New()
		clientPeer := cluster.NewPeer()
		serverPeer := cluster.NewPeer()
		server := testdata.NewTestServiceServer(srpc.NewServer(codec.JSON, srpc.WithLogger(logger.DefaulSLogger{})))
		defer server.Close()
		go server.Start(ctx, serverPeer.Listen())

		for b.Loop() {
			client := testdata.NewTestServiceClient(srpc.NewClient(serverPeer.Addr(), codec.JSON, clientPeer))
			resp, err := client.Add(ctx, testdata.AddReq{A: 10, B: 15})
			require.NoError(b, err)
			require.Equal(b, 25, resp.Result)
		}
	})

	b.Run("multiple clients parallel", func(b *testing.B) {
		cluster := New()
		clientPeer := cluster.NewPeer()
		serverPeer := cluster.NewPeer()
		server := testdata.NewTestServiceServer(srpc.NewServer(codec.JSON, srpc.WithLogger(logger.DefaulSLogger{})))
		defer server.Close()
		go server.Start(ctx, serverPeer.Listen())

		for b.Loop() {
			var wg sync.WaitGroup
			for range 100 {
				wg.Add(1)
				go func() {
					defer wg.Done()
					client := testdata.NewTestServiceClient(srpc.NewClient(serverPeer.Addr(), codec.JSON, clientPeer))
					resp, err := client.Add(ctx, testdata.AddReq{A: 10, B: 15})
					require.NoError(b, err)
					require.Equal(b, 25, resp.Result)
				}()
			}
			wg.Wait()
		}
	})

	b.Run("multiple clients parallel each multiple calls", func(b *testing.B) {
		cluster := New()
		clientPeer := cluster.NewPeer()
		serverPeer := cluster.NewPeer()
		server := testdata.NewTestServiceServer(srpc.NewServer(codec.JSON, srpc.WithLogger(logger.DefaulSLogger{})))
		defer server.Close()
		go server.Start(ctx, serverPeer.Listen())

		var wg sync.WaitGroup
		for b.Loop() {
			for range 10 {
				wg.Add(1)
				go func() {
					defer wg.Done()
					client := testdata.NewTestServiceClient(srpc.NewClient(serverPeer.Addr(), codec.JSON, clientPeer))
					for range 10 {
						resp, err := client.Add(ctx, testdata.AddReq{A: 10, B: 15})
						require.NoError(b, err)
						require.Equal(b, 25, resp.Result)
					}
				}()
			}
			wg.Wait()
		}
	})
}
