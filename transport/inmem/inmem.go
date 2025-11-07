package inmem

import (
	"context"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"math"
	"net"
	"sync"

	"github.com/tymbaca/srpc"
)

type Cluster struct {
	mu    sync.RWMutex
	peers map[string]*Peer

	lastID uint64
}

func New() *Cluster {
	return &Cluster{
		peers: make(map[string]*Peer),
	}
}

func (c *Cluster) NewPeer() *Peer {
	c.mu.Lock()
	defer c.mu.Unlock()

	addr := c.nextAddr()
	peer := &Peer{
		cluster: c, addr: addr,
		inbox: make(chan *conn),
	}

	c.peers[addr] = peer

	return peer
}

func (c *Cluster) getPeer(addr string) *Peer {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.peers[addr]
}

func (c *Cluster) nextAddr() string {
	if c.lastID == math.MaxUint64 {
		log.Panicf("inmem: max peer count (%d) reached", uint64(math.MaxUint64))
	}

	c.lastID++
	return fmt.Sprint(c.lastID)
}

type Peer struct {
	cluster *Cluster
	addr    string
	inbox   chan *conn
}

func (p *Peer) Listen() *PeerListener {
	l := &PeerListener{
		parent: p,
	}
	l.ctx, l.cancel = context.WithCancel(context.Background())
	return l
}

type PeerListener struct {
	parent *Peer
	ctx    context.Context
	cancel context.CancelFunc
}

// Accept waits and returns new connection to the listener.
// If Listener got closed Accept must return [ErrListenerClosed],
// including Accept calls that didn't returned yet.
func (pl *PeerListener) Accept() (srpc.Conn, error) {
	debug("wait for conn on inbox, peer: %+v", pl.parent)

	select {
	case <-pl.ctx.Done():
		return nil, srpc.ErrListenerClosed
	case conn := <-pl.parent.inbox:
		return conn, nil
	}
}

// Close closes the listener.
// Any blocked Accept operations will be unblocked and return errors.
// Close can be called multiple times.
func (pl *PeerListener) Close() error {
	pl.cancel()
	return nil
}

var ErrPeerNotFound = errors.New("peer not found")

func (p *Peer) Dial(ctx context.Context, addr string) (srpc.Conn, error) {
	target := p.cluster.getPeer(addr)
	if target == nil {
		return nil, ErrPeerNotFound
	}

	my, other := net.Pipe()

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case target.inbox <- &conn{
		c:          other,
		localAddr:  addr,
		remoteAddr: p.addr,
	}:
	}

	return &conn{
		c:          my,
		localAddr:  p.addr,
		remoteAddr: addr,
	}, nil
}

func (p *Peer) Addr() string {
	return p.addr
}

type conn struct {
	c          net.Conn
	localAddr  string
	remoteAddr string
}

func (c *conn) RemoteAddr() string {
	return c.remoteAddr
}

func (c *conn) LocalAddr() string {
	return c.localAddr
}

func (c *conn) Read(p []byte) (n int, err error) {
	return c.c.Read(p)
}

func (c *conn) Write(p []byte) (n int, err error) {
	return c.c.Write(p)
}

func (c *conn) Close() error {
	return c.c.Close()
}

const _debug = false

func debug(format string, args ...any) {
	if _debug {
		slog.Info(fmt.Sprintf(format, args...))
	}
}
