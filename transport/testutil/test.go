package testutil

import (
	"math/rand"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/tymbaca/srpc"
	"github.com/tymbaca/srpc/codec"
	"github.com/tymbaca/srpc/logger"
	"go.uber.org/goleak"
)

func TestSimple(t *testing.T, newListener func() srpc.Listener, newDialer func() srpc.Dialer) {
	ctx := t.Context()

	dialer := newDialer()
	listener := newListener()

	server := NewTestServiceServer(srpc.NewServer(codec.JSON))
	defer server.Close()
	go server.Start(ctx, listener)

	client := NewTestServiceClient(srpc.NewClient(listener.Addr(), codec.JSON, dialer))
	{
		resp, err := client.Add(ctx, AddReq{A: 10, B: 15})
		require.NoError(t, err)
		require.Equal(t, 25, resp.Result)
	}
	{
		resp, err := client.Divide(ctx, DivideReq{A: 10, B: 2})
		require.NoError(t, err)
		require.Equal(t, 5, resp.Result)
	}
	{
		_, err := client.Divide(ctx, DivideReq{A: 10, B: 0})
		require.Error(t, err)
	}
}

func TestStress(t *testing.T, newListener func() srpc.Listener, newDialer func() srpc.Dialer, clientCount, callPerClient int) {
	ctx := t.Context()
	defer goleak.VerifyNone(t)

	t.Run("single client", func(t *testing.T) {
		listener := newListener()
		server := NewTestServiceServer(srpc.NewServer(codec.JSON, srpc.WithLogger(logger.DefaulSLogger{})))
		defer server.Close()
		go server.Start(ctx, listener)

		client := NewTestServiceClient(srpc.NewClient(listener.Addr(), codec.JSON, newDialer()))
		resp, err := client.Add(ctx, AddReq{A: 10, B: 15})
		require.NoError(t, err)
		require.Equal(t, 25, resp.Result)
	})

	t.Run("multiple clients parallel each multiple calls", func(t *testing.T) {
		listener := newListener()
		server := NewTestServiceServer(srpc.NewServer(codec.JSON, srpc.WithLogger(logger.DefaulSLogger{})))
		defer server.Close()
		go server.Start(ctx, listener)

		var wg sync.WaitGroup
		for range clientCount {
			wg.Add(1)
			go func() {
				defer wg.Done()
				client := NewTestServiceClient(srpc.NewClient(listener.Addr(), codec.JSON, newDialer()))
				for range callPerClient {
					req := AddReq{A: rand.Int(), B: rand.Int()}
					resp, err := client.Add(ctx, req)
					require.NoError(t, err)
					require.Equal(t, req.A+req.B, resp.Result)
				}
			}()
		}
		wg.Wait()
	})

	t.Run("multiple clients parallel each multiple calls | close", func(t *testing.T) {
		listener := newListener()
		server := NewTestServiceServer(srpc.NewServer(codec.JSON, srpc.WithLogger(logger.DefaulSLogger{})))
		defer server.Close()
		go server.Start(ctx, listener)

		for range clientCount {
			go func() {
				client := NewTestServiceClient(srpc.NewClient(listener.Addr(), codec.JSON, newDialer()))
				for {
					req := AddReq{A: rand.Int(), B: rand.Int()}
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

func Benchmark(b *testing.B, newListener func() srpc.Listener, newDialer func() srpc.Dialer) {
	ctx := b.Context()
	dialer := newDialer()
	listener := newListener()

	server := NewTestServiceServer(srpc.NewServer(codec.JSON, srpc.WithLogger(logger.DefaulSLogger{})))
	defer server.Close()
	go server.Start(ctx, listener)

	client := NewTestServiceClient(srpc.NewClient(listener.Addr(), codec.JSON, dialer))

	for b.Loop() {
		req := AddReq{A: rand.Int(), B: rand.Int()}
		resp, err := client.Add(ctx, req)
		_, _ = resp, err
	}
}
