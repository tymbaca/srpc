// Package stls provides sRPC transport layer with key exchange and symmetric encription.
package stls

import (
	"crypto/ecdh"
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"io"

	"github.com/tymbaca/sbinary"
	"github.com/tymbaca/srpc"
	"golang.org/x/crypto/chacha20"
)

var _curve = ecdh.P256()

const (
	_keyLen        = 32
	_version uint8 = 1
)

func handshake(conn srpc.Conn, key *ecdh.PrivateKey, writeFirst bool) (_ *Conn, err error) {
	var hisPublicKey *ecdh.PublicKey
	if writeFirst {
		err = writeKey(conn, key.PublicKey())
		if err != nil {
			return nil, fmt.Errorf("write key: %w", err)
		}

		hisPublicKey, err = readKey(conn)
		if err != nil {
			return nil, fmt.Errorf("read key: %w", err)
		}
	} else {
		hisPublicKey, err = readKey(conn)
		if err != nil {
			return nil, fmt.Errorf("read key: %w", err)
		}

		err = writeKey(conn, key.PublicKey())
		if err != nil {
			return nil, fmt.Errorf("write key: %w", err)
		}
	}

	secretKey, err := key.ECDH(hisPublicKey)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, chacha20.NonceSizeX)
	rand.Read(nonce)

	stream, err := chacha20.NewUnauthenticatedCipher(secretKey, nonce)
	if err != nil {
		return nil, fmt.Errorf("create chacha20 cipher: %w", err)
	}

	return connWithCipher(conn, stream), nil
}

func writeKey(w io.Writer, key *ecdh.PublicKey) error {
	keyBytes := key.Bytes()
	msg := exchangeKeyMsg{
		Version: _version,
		Len:     uint8(len(keyBytes)),
		Key:     keyBytes,
	}
	return sbinary.NewEncoder(w).Encode(msg, binary.BigEndian)
}

func readKey(r io.Reader) (key *ecdh.PublicKey, err error) {
	var msg exchangeKeyMsg
	if err := sbinary.NewDecoder(r).Decode(&msg, binary.BigEndian); err != nil {
		return nil, err
	}

	key, err = ecdh.P256().NewPublicKey(msg.Key)
	if err != nil {
		return nil, fmt.Errorf("got incorrect public key from peer: %w", err)
	}

	return key, nil
}

type exchangeKeyMsg struct {
	Version uint8
	Len     uint8 `sbin:"lenof:Key"`
	Key     []byte
}
