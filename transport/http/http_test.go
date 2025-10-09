package httptransport

import (
	"math/rand/v2"
	"net/http"
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
	goleak.VerifyTestMain(m,
		goleak.IgnoreAnyFunction("net.(*sysDialer).dialParallel"),
		goleak.IgnoreAnyFunction("net.(*sysDialer).dialParallel.func1"),
		goleak.IgnoreAnyFunction("net.(*netFD).connect.func2"),
	)
}

func TestHttpTransport(t *testing.T) {
	ctx := t.Context()

	server := testdata.NewTestServiceServer(srpc.NewServer(codec.JSON))
	defer server.Close()
	go server.Start(ctx, CreateAndStartListener(":8080", "/srpc", http.MethodPost))

	client := testdata.NewTestServiceClient(srpc.NewClient("http://localhost:8080", codec.JSON, NewClientConnector("/srpc", http.MethodPost)))
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

func TestHttpTransportStress(t *testing.T) {
	ctx := t.Context()

	t.Run("single client", func(t *testing.T) {
		server := testdata.NewTestServiceServer(srpc.NewServer(codec.JSON, srpc.WithLogger(logger.DefaulSLogger{})))
		go server.Start(ctx, CreateAndStartListener(":8080", "/srpc", http.MethodPost))
		defer server.Close()

		client := testdata.NewTestServiceClient(srpc.NewClient("http://localhost:8080", codec.JSON, NewClientConnector("/srpc", http.MethodPost)))
		resp, err := client.Add(ctx, testdata.AddReq{A: 10, B: 15})
		require.NoError(t, err)
		require.Equal(t, 25, resp.Result)
	})

	t.Run("multiple clients parallel each multiple calls", func(t *testing.T) {
		t.Skip("flickery, now my fault")
		server := testdata.NewTestServiceServer(srpc.NewServer(codec.JSON, srpc.WithLogger(logger.DefaulSLogger{})))
		go server.Start(ctx, CreateAndStartListener(":8080", "/srpc", http.MethodPost))
		defer server.Close()

		var wg sync.WaitGroup
		for range 10 {
			wg.Add(1)
			go func() {
				defer wg.Done()
				client := testdata.NewTestServiceClient(srpc.NewClient("http://localhost:8080", codec.JSON, NewClientConnector("/srpc", http.MethodPost)))
				for range 10 {
					resp, err := client.Add(ctx, testdata.AddReq{A: 10, B: 15})
					require.NoError(t, err)
					require.Equal(t, 25, resp.Result)
				}
			}()
		}
		wg.Wait()
	})
}

func BenchmarkHttpTransportStress(b *testing.B) {
	ctx := b.Context()

	server := testdata.NewTestServiceServer(srpc.NewServer(codec.JSON, srpc.WithLogger(logger.DefaulSLogger{})))
	go server.Start(ctx, CreateAndStartListener(":8080", "/srpc", http.MethodPost))
	defer server.Close()

	client := testdata.NewTestServiceClient(srpc.NewClient("http://localhost:8080", codec.JSON, NewClientConnector("/srpc", http.MethodPost)))

	for b.Loop() {
		req := testdata.AddReq{A: rand.Int(), B: rand.Int()}
		resp, err := client.Add(ctx, req)
		_, _ = resp, err
	}
}
