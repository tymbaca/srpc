package stdnet

import (
	"testing"

	"github.com/tymbaca/srpc"
	"github.com/tymbaca/srpc/transport/testutil"
	"go.uber.org/goleak"
)

func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m)
}

const addr = ":6666"

func TestSimple(t *testing.T) {
	t.Run("tcp", func(t *testing.T) { testutil.TestSimple(t, newListenerFunc("tcp", addr), newDialerFunc("tcp")) })
	t.Run("udp", func(t *testing.T) { testutil.TestSimple(t, newListenerFunc("udp", addr), newDialerFunc("udp")) })
	t.Run("unix", func(t *testing.T) { testutil.TestSimple(t, newListenerFunc("unix", addr), newDialerFunc("unix")) })
}

func TestStress(t *testing.T) {
	t.Run("tcp", func(t *testing.T) {
		testutil.TestStress(t, newListenerFunc("tcp", addr), newDialerFunc("tcp"), 100, 100)
	})
	t.Run("udp", func(t *testing.T) {
		testutil.TestStress(t, newListenerFunc("udp", addr), newDialerFunc("udp"), 100, 100)
	})
	t.Run("unix", func(t *testing.T) {
		testutil.TestStress(t, newListenerFunc("unix", addr), newDialerFunc("unix"), 100, 100)
	})
}

func BenchmarkStress(b *testing.B) {
	b.Run("tcp", func(b *testing.B) { testutil.Benchmark(b, newListenerFunc("tcp", addr), newDialerFunc("tcp")) })
	b.Run("udp", func(b *testing.B) { testutil.Benchmark(b, newListenerFunc("udp", addr), newDialerFunc("udp")) })
	b.Run("unix", func(b *testing.B) { testutil.Benchmark(b, newListenerFunc("unix", addr), newDialerFunc("unix")) })
}

func newListenerFunc(network string, addr string) func() srpc.Listener {
	return func() srpc.Listener {
		l, err := Listen(network, addr)
		if err != nil {
			panic(err)
		}
		return l
	}
}

func newDialerFunc(network string) func() srpc.Dialer {
	return func() srpc.Dialer {
		return NewDialer(network)
	}
}
