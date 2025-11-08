package stls

import (
	"crypto/ecdh"
	"crypto/rand"
	"fmt"

	"github.com/tymbaca/srpc"
)

func NewListenerRandomKey(backing srpc.Listener) (*Listener, error) {
	key, err := _curve.GenerateKey(rand.Reader)
	if err != nil {
		return nil, err
	}

	return NewListener(backing, key)
}

func NewListener(backing srpc.Listener, key *ecdh.PrivateKey) (*Listener, error) {
	if key.Curve() != _curve {
		return nil, fmt.Errorf("invalid key curve: got %s, must be %s", key.Curve(), _curve)
	}

	return &Listener{
		backing: backing,
		key:     key,
	}, nil
}

type Listener struct {
	backing srpc.Listener
	key     *ecdh.PrivateKey
}

// Accept waits and returns new connection to the listener.
// If Listener got closed Accept must return [ErrListenerClosed],
// including Accept calls that didn't returned yet.
func (l *Listener) Accept() (srpc.Conn, error) {
	conn, err := l.backing.Accept()
	if err != nil {
		return nil, err
	}

	return handshake(conn, l.key, true)
}

// Close closes the listener.
// Any blocked Accept operations will be unblocked and return errors.
// Close can be called multiple times.
func (l *Listener) Close() error {
	return l.backing.Close()
}

// Addr returns listener's address. Address is valid to use in [Dialer.Dial].
func (l *Listener) Addr() string {
	return l.backing.Addr()
}
