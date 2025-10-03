package httptransport

import (
	"net/http"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tymbaca/srpc"
	"github.com/tymbaca/srpc/codec"
	"github.com/tymbaca/srpc/logger"
	"github.com/tymbaca/srpc/transport/http/testdata"
	"go.uber.org/goleak"
)

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

func BenchmarkHttpTransport(b *testing.B) {
	defer goleak.VerifyNone(b, goleak.IgnoreCurrent())
	ctx := b.Context()

	b.Run("single client", func(b *testing.B) {
		server := testdata.NewTestServiceServer(srpc.NewServer(codec.JSON, srpc.WithLogger(logger.DefaulSLogger{})))
		go server.Start(ctx, CreateAndStartListener(":8080", "/srpc", http.MethodPost))
		defer server.Close()

		client := testdata.NewTestServiceClient(srpc.NewClient("http://localhost:8080", codec.JSON, NewClientConnector("/srpc", http.MethodPost)))
		for b.Loop() {
			resp, err := client.Add(ctx, testdata.AddReq{A: 10, B: 15})
			require.NoError(b, err)
			require.Equal(b, 25, resp.Result)
		}
	})

	b.Run("multiple clients sequentual", func(b *testing.B) {
		server := testdata.NewTestServiceServer(srpc.NewServer(codec.JSON, srpc.WithLogger(logger.DefaulSLogger{})))
		go server.Start(ctx, CreateAndStartListener(":8080", "/srpc", http.MethodPost))
		defer server.Close()

		for b.Loop() {
			client := testdata.NewTestServiceClient(srpc.NewClient("http://localhost:8080", codec.JSON, NewClientConnector("/srpc", http.MethodPost)))
			resp, err := client.Add(ctx, testdata.AddReq{A: 10, B: 15})
			require.NoError(b, err)
			require.Equal(b, 25, resp.Result)
		}
	})

	b.Run("multiple clients parallel", func(b *testing.B) {
		server := testdata.NewTestServiceServer(srpc.NewServer(codec.JSON, srpc.WithLogger(logger.DefaulSLogger{})))
		go server.Start(ctx, CreateAndStartListener(":8080", "/srpc", http.MethodPost))
		defer server.Close()

		for b.Loop() {
			var wg sync.WaitGroup
			for range 1000 {
				wg.Add(1)
				go func() {
					defer wg.Done()
					client := testdata.NewTestServiceClient(srpc.NewClient("http://localhost:8080", codec.JSON, NewClientConnector("/srpc", http.MethodPost)))
					resp, err := client.Add(ctx, testdata.AddReq{A: 10, B: 15})
					require.NoError(b, err)
					require.Equal(b, 25, resp.Result)
				}()
			}
			wg.Wait()
		}
	})

	b.Run("multiple clients parallel each multiple calls", func(b *testing.B) {
		server := testdata.NewTestServiceServer(srpc.NewServer(codec.JSON, srpc.WithLogger(logger.DefaulSLogger{})))
		go server.Start(ctx, CreateAndStartListener(":8080", "/srpc", http.MethodPost))
		defer server.Close()

		for b.Loop() {
			var wg sync.WaitGroup
			for range 10 {
				wg.Add(1)
				go func() {
					defer wg.Done()
					client := testdata.NewTestServiceClient(srpc.NewClient("http://localhost:8080", codec.JSON, NewClientConnector("/srpc", http.MethodPost)))
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
