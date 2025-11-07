package inmem

import (
	"testing"

	"github.com/tymbaca/srpc"
	"github.com/tymbaca/srpc/transport/testutil"
	"go.uber.org/goleak"
)

func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m)
}

func TestSimple(t *testing.T) {
	cluster := New()
	testutil.TestSimple(t,
		func() srpc.Listener { return cluster.NewPeer().Listen() },
		func() srpc.Dialer { return cluster.NewPeer() },
	)
}

func TestStress(t *testing.T) {
	cluster := New()
	testutil.TestStress(t,
		func() srpc.Listener { return cluster.NewPeer().Listen() },
		func() srpc.Dialer { return cluster.NewPeer() },
		100,
		100,
	)
}

func BenchmarkStress(b *testing.B) {
	cluster := New()
	testutil.Benchmark(b,
		func() srpc.Listener { return cluster.NewPeer().Listen() },
		func() srpc.Dialer { return cluster.NewPeer() },
	)
}
