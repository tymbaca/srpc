package inmem

import (
	"context"
	"errors"
	"fmt"
	"log"
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
	peer := &Peer{cluster: c, addr: addr}
	peer.peerListener.parent = peer // TODO: do i need this?

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

	*peerListener
}

type peerListener struct {
	parent *Peer
	ch     chan req
}

type req struct {
	replyCh chan resp
}

type resp struct {
}

// Accept waits and returns new connection to the listener.
// If Listener got closed Accept must return [ErrListenerClosed],
// including Accept calls that didn't returned yet.
func (pl *peerListener) Accept() (srpc.ServerConn, error) {
	pl.ch
}

// Close closes the listener.
// Any blocked Accept operations will be unblocked and return errors.
// Close can be called multiple times.
func (pl *peerListener) Close() error {

}

var ErrPeerNotFound = errors.New("peer not found")

func (p *Peer) Connect(ctx context.Context, addr string) (srpc.ClientConn, error) {
	target := p.cluster.getPeer(addr)
	if target == nil {
		return nil, ErrPeerNotFound
	}

	// return conn{p, target}, nil
	panic("not implemented")
}

func (p *Peer) Addr() string {
	return p.addr
}

//
// type conn struct {
// 	peerA, peerB *Peer
// }
//
// func (co *conn) Send(ctx context.Context, req srpc.Request) (srpc.Response, error) {
// 	panic("not implemented") // TODO: Implement
// }
//
// // Close must be called after Send
// func (co *conn) Close() error {
// 	panic("not implemented") // TODO: Implement
// }
