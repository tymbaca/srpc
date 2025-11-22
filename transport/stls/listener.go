package stls

import (
	"crypto/ecdh"
	_ "crypto/rand"
	"errors"
	"fmt"
	"io"

	transport "github.com/tymbaca/srpc/transport"
)

// NewListenerRandomKey is the same as [NewListener] but it generates a key
// using provided random reader (e.g. [rand.Reader]).
func NewListenerRandomKey(backing transport.Listener, rand io.Reader) (*Listener, error) {
	key, err := _curve.GenerateKey(rand)
	if err != nil {
		return nil, err
	}

	return NewListener(backing, key)
}

// NewListener creates new stls listener with provided backing listener
// and a private key.
func NewListener(backing transport.Listener, key *ecdh.PrivateKey) (*Listener, error) {
	if key.Curve() != _curve {
		return nil, fmt.Errorf("invalid key curve: got %s, must be %s", key.Curve(), _curve)
	}

	return &Listener{
		backing: backing,
		key:     key,
	}, nil
}

// Listener provides stls security layer of the backing listener.
type Listener struct {
	backing transport.Listener
	key     *ecdh.PrivateKey
}

// Accept waits and returns new connection to the listener.
// If Listener got closed Accept must return [ErrListenerClosed],
// including Accept calls that didn't returned yet.
func (l *Listener) Accept() (transport.Conn, error) {
	conn, err := l.backing.Accept()
	if err != nil {
		return nil, err
	}

	stlsConn, err := handshake(conn, l.key, true)
	if err != nil {
		if closeErr := conn.Close(); closeErr != nil {
			err = errors.Join(err, closeErr)
		}
		return nil, err
	}

	return stlsConn, nil
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
