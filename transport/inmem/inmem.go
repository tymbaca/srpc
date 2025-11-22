// Package inmem provides in-memory srpc transport implementation.
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

	"github.com/tymbaca/srpc/transport"
)

// ErrPeerNotFound is returned in [Peer.Dial] if there is no peer with specified address.
var ErrPeerNotFound = errors.New("peer not found")

// Cluster represents a collection of [Peer]s that can connect to each other.
type Cluster struct {
	mu    sync.RWMutex
	peers map[string]*Peer

	lastID uint64
}

// New creates new [Cluster].
func New() *Cluster {
	return &Cluster{
		peers: make(map[string]*Peer),
	}
}

// NewPeer creates new [Peer] with unique address.
// The addres is an unsigned integer that gets incremented on each call.
// It panics if too much peers was created (more then math.MaxUint64).
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

// Peer represents a single node the cluster.
// It can act as both a [transport.Dialer] and a [transport.Listener].
type Peer struct {
	cluster *Cluster
	addr    string
	inbox   chan *conn
}

// Listen creates a new [transport.Listener]. It's callers responsibily to close the listener.
func (p *Peer) Listen() transport.Listener {
	l := &peerListener{
		parent: p,
	}
	l.ctx, l.cancel = context.WithCancel(context.Background())
	return l
}

type peerListener struct {
	parent *Peer
	ctx    context.Context
	cancel context.CancelFunc
}

// Accept implements [transport.Listener].
func (pl *peerListener) Accept() (transport.Conn, error) {
	debug("wait for conn on inbox, peer: %+v", pl.parent)

	select {
	case <-pl.ctx.Done():
		return nil, transport.ErrListenerClosed
	case conn := <-pl.parent.inbox:
		return conn, nil
	}
}

// Close implements [transport.Listener].
func (pl *peerListener) Close() error {
	pl.cancel()
	return nil
}

// Addr implements [transport.Listener].
func (pl *peerListener) Addr() string {
	return pl.parent.addr
}

// Dial implements [transport.Dialer]
func (p *Peer) Dial(ctx context.Context, addr string) (transport.Conn, error) {
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

// Addr returns the address of current peer.
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
