package stls

import (
	"context"
	"crypto/ecdh"
	"crypto/rand"
	"fmt"

	"github.com/tymbaca/srpc"
)

func NewDialerRandomKey(backing srpc.Dialer) (*Dialer, error) {
	key, err := _curve.GenerateKey(rand.Reader)
	if err != nil {
		return nil, err
	}

	return NewDialer(backing, key)
}

func NewDialer(backing srpc.Dialer, key *ecdh.PrivateKey) (*Dialer, error) {
	if key.Curve() != _curve {
		return nil, fmt.Errorf("invalid key curve: got %s, must be %s", key.Curve(), _curve)
	}

	return &Dialer{
		backing: backing,
		key:     key,
	}, nil
}

type Dialer struct {
	backing srpc.Dialer
	key     *ecdh.PrivateKey
}

func (d *Dialer) Dial(ctx context.Context, addr string) (srpc.Conn, error) {
	conn, err := d.backing.Dial(ctx, addr)
	if err != nil {
		return nil, err
	}

	return handshake(conn, d.key, false)
}
