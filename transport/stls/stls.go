// Package stls provides sRPC transport layer with key exchange and symmetric encryption.
package stls

import (
	"bytes"
	"crypto/ecdh"
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"io"

	"github.com/tymbaca/sbinary"
	"github.com/tymbaca/srpc/transport"
	"golang.org/x/crypto/chacha20"
)

var _curve = ecdh.P256()

const (
	_keyLen        = 32
	_version uint8 = 1
)

type exchangeKeyMsg struct {
	Version  uint8
	KeyLen   uint8 `sbin:"lenof:Key"`
	Key      []byte
	NonceLen uint8 `sbin:"lenof:Nonce"`
	Nonce    []byte
}

func handshake(conn transport.Conn, key *ecdh.PrivateKey, lead bool) (_ *Conn, err error) {
	var (
		hisPublicKey *ecdh.PublicKey
		nonce        []byte
	)
	if lead {
		nonce = make([]byte, chacha20.NonceSizeX)
		rand.Read(nonce)

		err = writeMsg(conn, key.PublicKey(), nonce)
		if err != nil {
			return nil, fmt.Errorf("write key: %w", err)
		}

		var gotNonce []byte
		hisPublicKey, gotNonce, err = readKey(conn)
		if err != nil {
			return nil, fmt.Errorf("read key: %w", err)
		}
		if !bytes.Equal(nonce, gotNonce) {
			return nil, fmt.Errorf("peer sent incorrect nonce: got %v, want %v", gotNonce, nonce)
		}
	} else {
		hisPublicKey, nonce, err = readKey(conn)
		if err != nil {
			return nil, fmt.Errorf("read key: %w", err)
		}

		err = writeMsg(conn, key.PublicKey(), nonce)
		if err != nil {
			return nil, fmt.Errorf("write key: %w", err)
		}
	}

	secretKey, err := key.ECDH(hisPublicKey)
	if err != nil {
		return nil, err
	}

	stream, err := chacha20.NewUnauthenticatedCipher(secretKey, nonce)
	if err != nil {
		return nil, fmt.Errorf("create chacha20 cipher: %w", err)
	}

	return connWithCipher(conn, stream), nil
}

func writeMsg(w io.Writer, key *ecdh.PublicKey, nonce []byte) error {
	keyBytes := key.Bytes()
	msg := exchangeKeyMsg{
		Version:  _version,
		KeyLen:   uint8(len(keyBytes)),
		Key:      keyBytes,
		NonceLen: uint8(len(nonce)),
		Nonce:    nonce,
	}
	return sbinary.NewEncoder(w).Encode(msg, binary.BigEndian)
}

func readKey(r io.Reader) (key *ecdh.PublicKey, nonce []byte, err error) {
	var msg exchangeKeyMsg
	if err := sbinary.NewDecoder(r).Decode(&msg, binary.BigEndian); err != nil {
		return nil, nil, err
	}

	key, err = ecdh.P256().NewPublicKey(msg.Key)
	if err != nil {
		return nil, nil, fmt.Errorf("got incorrect public key from peer: %w", err)
	}

	return key, msg.Nonce, nil
}
