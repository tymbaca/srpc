package stls

import (
	"context"
	"crypto/ecdh"
	"crypto/rand"
	"fmt"

	transport "github.com/tymbaca/srpc/transport"
)

// NewDialerRandomKey is the same as [NewDialer] but it generates a key
// using provided random reader (e.g. [rand.Reader]).
func NewDialerRandomKey(backing transport.Dialer) (*Dialer, error) {
	key, err := _curve.GenerateKey(rand.Reader)
	if err != nil {
		return nil, err
	}

	return NewDialer(backing, key)
}

// NewDialer creates new stls dialer with provided backing dialer
// and a private key.
func NewDialer(backing transport.Dialer, key *ecdh.PrivateKey) (*Dialer, error) {
	if key.Curve() != _curve {
		return nil, fmt.Errorf("invalid key curve: got %s, must be %s", key.Curve(), _curve)
	}

	return &Dialer{
		backing: backing,
		key:     key,
	}, nil
}

// Dialer provides stls security layer of the backing dialer.
type Dialer struct {
	backing transport.Dialer
	key     *ecdh.PrivateKey
}

// Dial implemetns [transport.Dialer]
func (d *Dialer) Dial(ctx context.Context, addr string) (transport.Conn, error) {
	conn, err := d.backing.Dial(ctx, addr)
	if err != nil {
		return nil, err
	}

	return handshake(conn, d.key, false)
}
