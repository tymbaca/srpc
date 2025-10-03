package httptransport

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tymbaca/srpc"
	"github.com/tymbaca/srpc/codec"
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
		server := testdata.NewTestServiceServer(srpc.NewServer(codec.JSON))
		defer server.Close()
		go server.Start(ctx, CreateAndStartListener(":8080", "/srpc", http.MethodPost))

		client := testdata.NewTestServiceClient(srpc.NewClient("http://localhost:8080", codec.JSON, NewClientConnector("/srpc", http.MethodPost)))
		for b.Loop() {
			resp, err := client.Add(ctx, testdata.AddReq{A: 10, B: 15})
			require.NoError(b, err)
			require.Equal(b, 25, resp.Result)
		}
	})
}
