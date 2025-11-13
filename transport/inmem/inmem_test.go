package inmem

import (
	"testing"

	"github.com/tymbaca/srpc/transport"
	"github.com/tymbaca/srpc/transport/testutil"
)

func TestSimple(t *testing.T) {
	cluster := New()
	testutil.TestSimple(t,
		func() transport.Listener { return cluster.NewPeer().Listen() },
		func() transport.Dialer { return cluster.NewPeer() },
	)
}

func TestStress(t *testing.T) {
	cluster := New()
	testutil.TestComplex(t,
		func() transport.Listener { return cluster.NewPeer().Listen() },
		func() transport.Dialer { return cluster.NewPeer() },
		100,
		100,
	)
}

func BenchmarkStress(b *testing.B) {
	cluster := New()
	testutil.Benchmark(b,
		func() transport.Listener { return cluster.NewPeer().Listen() },
		func() transport.Dialer { return cluster.NewPeer() },
	)
}
