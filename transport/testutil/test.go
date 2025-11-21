// Package testutil provides common tests and utilities for transport implementations.
package testutil

import (
	"context"
	cryptoRand "crypto/rand"
	"fmt"
	"math/rand"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/tymbaca/srpc"
	"github.com/tymbaca/srpc/call"
	"github.com/tymbaca/srpc/codec/json"
	"github.com/tymbaca/srpc/logger"
	"github.com/tymbaca/srpc/metadata"
	"github.com/tymbaca/srpc/status"
	"github.com/tymbaca/srpc/transport"
	"go.uber.org/goleak"
)

func goleakOpts() []goleak.Option {
	return []goleak.Option{
		goleak.IgnoreCurrent(),
		goleak.IgnoreAnyFunction("github.com/tymbaca/srpc/transport/testutil.(*TestServiceServer).LongAdd"),
	}
}

func TestSimple(t *testing.T, newListener func() transport.Listener, newDialer func() transport.Dialer) {
	defer goleak.VerifyNone(t, goleakOpts()...)
	ctx := t.Context()

	dialer := newDialer()
	listener := newListener()

	server := NewTestServiceServer(srpc.NewServer(json.Codec, srpc.WithConnErrorHandler(func(err error) error {
		panic(err)
		// return nil // TODO: undo
	})))
	defer server.Close()
	go server.Start(ctx, listener)

	client := NewTestServiceClient(srpc.NewClient(listener.Addr(), json.Codec, dialer))
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
		code, ok := status.FromError(err)
		require.True(t, ok)
		require.Equal(t, status.InvalidArgument, code)
	}
	{
		_, err := client.Divide(ctx, DivideReq{A: 10, B: -1})
		require.Error(t, err)
		code, ok := status.FromError(err)
		require.True(t, ok)
		require.Equal(t, status.ErrorFromService, code)
	}
	{
		ctx := metadata.ToContext(ctx, metadata.Metadata{
			"k1": {"v1", "v2"},
		})

		resp, err := client.ReplyMD(ctx, ReplyMDReq{Key: "k1"})
		require.NoError(t, err)
		require.Equal(t, ReplyMDResp{Vals: []string{"v1", "v2"}, Ok: true}, resp)

		var respMD metadata.Metadata
		ctx = call.WithOptions(ctx, call.WithResponseMetadata(&respMD))
		resp, err = client.ReplyMD(ctx, ReplyMDReq{Key: "k1", RespMDKey: "rk1", RespMDVals: []string{"rv1", "rv2"}})
		require.NoError(t, err)
		require.Equal(t, ReplyMDResp{Vals: []string{"v1", "v2"}, Ok: true}, resp)
		require.Equal(t, metadata.Metadata(map[string][]string{"rk1": {"rv1", "rv2"}}), respMD)

		resp, err = client.ReplyMD(ctx, ReplyMDReq{Key: "badkey"})
		require.NoError(t, err)
		require.Equal(t, ReplyMDResp{Ok: false}, resp)
	}
	{
		input := make([]byte, 1024*1024)
		cryptoRand.Read(input)

		resp, err := client.Blob(ctx, Blob{Data: input})
		require.NoError(t, err)
		require.Equal(t, input, resp.Data)
	}
}

func TestComplex(t *testing.T, newListener func() transport.Listener, newDialer func() transport.Dialer, clientCount, callPerClient int) {
	defer goleak.VerifyNone(t, goleakOpts()...)
	ctx := t.Context()

	t.Run("single client", func(t *testing.T) {
		listener := newListener()
		server := NewTestServiceServer(srpc.NewServer(json.Codec, srpc.WithLogger(logger.DefaultSLogger{})))
		defer server.Close()
		go server.Start(ctx, listener)

		client := NewTestServiceClient(srpc.NewClient(listener.Addr(), json.Codec, newDialer()))
		resp, err := client.Add(ctx, AddReq{A: 10, B: 15})
		require.NoError(t, err)
		require.Equal(t, 25, resp.Result)
	})

	t.Run("close conn after successful dial (no goroutines must be left)", func(t *testing.T) {
		listener := newListener()
		server := NewTestServiceServer(srpc.NewServer(json.Codec, srpc.WithLogger(logger.DefaultSLogger{})))
		defer server.Close()
		go server.Start(ctx, listener)

		//
		d, cancel := newInterruptDialer(ctx, newDialer(), &interruptConfig{CloseAfterDial: true})
		defer cancel()

		client := NewTestServiceClient(srpc.NewClient(listener.Addr(), json.Codec, d))

		// NOTE: we need big input to exceed the [chunked.BufferedWriter] buffer,
		// because if we use smaller input, then we won't catch leaking
		// goroutine from [pipe.ToReader] (it would fully write data
		// into the reading side buffer and exit).
		// See commit afc7710ae76fe85bfb83296884f345c89857e9e2 for details.
		input := make([]byte, 1024*1024)
		cryptoRand.Read(input)
		_, err := client.Blob(ctx, Blob{Data: input})
		require.Error(t, err)
	})

	t.Run("single client, different methods", func(t *testing.T) {
		TestSimple(t, newListener, newDialer)
	})

	ctxCancelErrMsg := func(wait, waitCheck, dur time.Duration) string {
		return fmt.Sprintf("exit after context cancelation was too long, ctx canceled after %s, function took %s to exit (limit was %s)", wait, dur-wait, waitCheck-wait)
	}

	t.Run("single client, context check", func(t *testing.T) {
		listener := newListener()
		server := NewTestServiceServer(srpc.NewServer(json.Codec, srpc.WithLogger(logger.DefaultSLogger{})))
		defer server.Close()
		go server.Start(ctx, listener)

		client := NewTestServiceClient(srpc.NewClient(listener.Addr(), json.Codec, newDialer()))
		wait := 20 * time.Millisecond
		waitCheck := 25 * time.Millisecond
		t.Run("timeout", func(t *testing.T) {
			ctx, cancel := context.WithTimeout(ctx, wait)
			defer cancel()

			start := time.Now()
			_, err := client.LongAdd(ctx, AddReq{A: 10, B: 15})
			dur := time.Since(start)
			require.Error(t, err)
			require.Less(t, dur, waitCheck, ctxCancelErrMsg(wait, waitCheck, dur))
		})

		t.Run("cancel", func(t *testing.T) {
			ctx, cancel := context.WithCancel(ctx)
			defer cancel()

			go func() {
				select {
				case <-ctx.Done():
				case <-time.After(wait):
					cancel()
				}
			}()

			start := time.Now()
			_, err := client.LongAdd(ctx, AddReq{A: 10, B: 15})
			dur := time.Since(start)

			require.Error(t, err)
			require.Less(t, dur, waitCheck, ctxCancelErrMsg(wait, waitCheck, dur))
		})
		t.Run("already closed", func(t *testing.T) {
			ctx, cancel := context.WithCancel(ctx)
			cancel()

			_, err := client.LongAdd(ctx, AddReq{A: 10, B: 15})
			require.Error(t, err)
		})
	})

	t.Run("server, context check", func(t *testing.T) {
		server := NewTestServiceServer(srpc.NewServer(json.Codec, srpc.WithLogger(logger.DefaultSLogger{})))
		defer server.Close()

		wait := 20 * time.Millisecond
		waitCheck := 25 * time.Millisecond
		t.Run("timeout", func(t *testing.T) {
			ctx, cancel := context.WithTimeout(ctx, wait)
			defer cancel()

			start := time.Now()
			err := server.Start(ctx, newListener())
			dur := time.Since(start)
			require.NoError(t, err)
			require.Less(t, dur, waitCheck, ctxCancelErrMsg(wait, waitCheck, dur))
		})
		t.Run("cancel", func(t *testing.T) {
			ctx, cancel := context.WithCancel(ctx)
			defer cancel()

			go func() {
				select {
				case <-ctx.Done():
				case <-time.After(wait):
					cancel()
				}
			}()

			start := time.Now()
			err := server.Start(ctx, newListener())
			dur := time.Since(start)
			require.NoError(t, err)
			require.Less(t, dur, waitCheck, ctxCancelErrMsg(wait, waitCheck, dur))
		})
		t.Run("already canceled", func(t *testing.T) {
			ctx, cancel := context.WithCancel(ctx)
			cancel()

			err := server.Start(ctx, newListener())
			require.NoError(t, err)
		})
	})

	t.Run("multiple clients parallel each multiple calls", func(t *testing.T) {
		listener := newListener()
		server := NewTestServiceServer(srpc.NewServer(json.Codec, srpc.WithLogger(logger.DefaultSLogger{})))
		defer server.Close()
		go server.Start(ctx, listener)

		var wg sync.WaitGroup
		for range clientCount {
			wg.Add(1)
			go func() {
				defer wg.Done()
				client := NewTestServiceClient(srpc.NewClient(listener.Addr(), json.Codec, newDialer()))
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
		server := NewTestServiceServer(srpc.NewServer(json.Codec, srpc.WithLogger(logger.DefaultSLogger{})))
		defer server.Close()
		go server.Start(ctx, listener)

		for range clientCount {
			go func() {
				client := NewTestServiceClient(srpc.NewClient(listener.Addr(), json.Codec, newDialer()))
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

func Benchmark(b *testing.B, newListener func() transport.Listener, newDialer func() transport.Dialer) {
	ctx := b.Context()
	dialer := newDialer()
	listener := newListener()

	server := NewTestServiceServer(srpc.NewServer(json.Codec, srpc.WithLogger(logger.DefaultSLogger{})))
	defer server.Close()
	go server.Start(ctx, listener)

	client := NewTestServiceClient(srpc.NewClient(listener.Addr(), json.Codec, dialer))

	for b.Loop() {
		req := AddReq{A: rand.Int(), B: rand.Int()}
		resp, err := client.Add(ctx, req)
		_, _ = resp, err
	}
}
