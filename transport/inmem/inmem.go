package inmem

import (
	"context"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"math"
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
	inbox   chan *conn // only for delegating to peerListener
}

func (p *Peer) Listen() *peerListener {
	l := &peerListener{
		parent: p,
		inbox:  p.inbox,
	}
	l.ctx, l.cancel = context.WithCancel(context.Background())
	return l
}

type peerListener struct {
	parent *Peer // for debug purposes
	ctx    context.Context
	cancel context.CancelFunc

	inbox chan *conn
}

// Accept waits and returns new connection to the listener.
// If Listener got closed Accept must return [ErrListenerClosed],
// including Accept calls that didn't returned yet.
func (pl *peerListener) Accept() (srpc.ServerConn, error) {
	debug("wait for conn on inbox, peer: %+v", pl.parent)

	select {
	case <-pl.ctx.Done():
		return nil, srpc.ErrListenerClosed
	case conn := <-pl.inbox:
		return conn, nil
	}
}

// Close closes the listener.
// Any blocked Accept operations will be unblocked and return errors.
// Close can be called multiple times.
func (pl *peerListener) Close() error {
	pl.cancel()
	return nil
}

var ErrPeerNotFound = errors.New("peer not found")

func (p *Peer) Connect(_ context.Context, addr string) (srpc.ClientConn, error) {
	target := p.cluster.getPeer(addr)
	if target == nil {
		return nil, ErrPeerNotFound
	}

	return &conn{
		client: p, server: target,
	}, nil
}

func (p *Peer) Addr() string {
	return p.addr
}

type conn struct {
	client, server *Peer

	ctx     context.Context
	cancel  context.CancelFunc
	req     srpc.Request
	replyCh chan srpc.Response
}

func (c *conn) Do(ctx context.Context, req srpc.Request) (srpc.Response, error) {
	c.req = req
	c.replyCh = make(chan srpc.Response)
	c.ctx, c.cancel = context.WithCancel(ctx)
	defer c.cancel()

	debug("wait send conn to target inbox, me: %+v, target: %+v", c.client, c.server)

	select {
	case <-c.ctx.Done():
		return srpc.Response{}, ctx.Err()
	case c.server.inbox <- c:
	}

	select {
	case <-c.ctx.Done():
		return srpc.Response{}, ctx.Err()
	case resp := <-c.replyCh:
		return resp, nil
	}
}

func (c *conn) Request() srpc.Request {
	return c.req
}

func (c *conn) Addr() string {
	return c.client.Addr()
}

func (c *conn) Reply(ctx context.Context, resp srpc.Response) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case c.replyCh <- resp:
		return nil
	}
}

// Close must be called after Send
func (c *conn) Close() error {
	if c.cancel != nil {
		c.cancel()
	}

	return nil
}

const _debug = false

func debug(format string, args ...any) {
	if _debug {
		slog.Info(fmt.Sprintf(format, args...))
	}
}
