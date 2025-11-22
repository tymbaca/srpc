// Package stls provides srpc transport layer with key exchange and symmetric encryption.
// It wraps another backing transport implementation.
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
	_version uint8 = 1
)

var _proto = [...]byte{'s', 't', 'l', 's'}

type exchangeKeyMsg struct {
	Version  uint8   // currently no-op, always 1
	Seq      uint32  // leader peer randomly generates it. second peer increments it by 1 and sends it back
	Proto    [4]byte // always "stls"
	KeyLen   uint8   `sbin:"lenof:Key"`
	Key      []byte
	NonceLen uint8 `sbin:"lenof:Nonce"`
	Nonce    []byte
}

func handshake(conn transport.Conn, key *ecdh.PrivateKey, lead bool) (_ *conn, err error) {
	var (
		hisPublicKey *ecdh.PublicKey
		nonce        []byte
		seq          uint32
	)
	if lead {
		nonce = make([]byte, chacha20.NonceSizeX)
		rand.Read(nonce)

		seq = randSeq()

		err = writeMsg(conn, key.PublicKey(), nonce, seq)
		if err != nil {
			return nil, fmt.Errorf("write key: %w", err)
		}

		var gotNonce []byte
		var otherSeq uint32
		hisPublicKey, gotNonce, otherSeq, err = readKey(conn)
		if err != nil {
			return nil, fmt.Errorf("read key: %w", err)
		}
		if otherSeq != seq+1 {
			return nil, fmt.Errorf("peer sent incorrect seq: got %d, want %d+1", otherSeq, seq)
		}
		if !bytes.Equal(nonce, gotNonce) {
			return nil, fmt.Errorf("peer sent incorrect nonce: got %v, want %v", gotNonce, nonce)
		}
	} else {
		hisPublicKey, nonce, seq, err = readKey(conn)
		if err != nil {
			return nil, fmt.Errorf("read key: %w", err)
		}

		err = writeMsg(conn, key.PublicKey(), nonce, seq+1)
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

func randSeq() uint32 {
	var seq [4]byte
	rand.Read(seq[:])
	return binary.NativeEndian.Uint32(seq[:])
}

func writeMsg(w io.Writer, key *ecdh.PublicKey, nonce []byte, seq uint32) error {
	keyBytes := key.Bytes()
	msg := exchangeKeyMsg{
		Version:  _version,
		Seq:      seq,
		Proto:    _proto,
		KeyLen:   uint8(len(keyBytes)),
		Key:      keyBytes,
		NonceLen: uint8(len(nonce)),
		Nonce:    nonce,
	}
	return sbinary.NewEncoder(w).Encode(msg, binary.BigEndian)
}

func readKey(r io.Reader) (key *ecdh.PublicKey, nonce []byte, seq uint32, err error) {
	var msg exchangeKeyMsg
	if err := sbinary.NewDecoder(r).Decode(&msg, binary.BigEndian); err != nil {
		return nil, nil, 0, err
	}

	if msg.Proto != _proto {
		return nil, nil, 0, fmt.Errorf("peer protocol is not sTLS")
	}

	key, err = ecdh.P256().NewPublicKey(msg.Key)
	if err != nil {
		return nil, nil, 0, fmt.Errorf("got incorrect public key from peer: %w", err)
	}

	return key, msg.Nonce, msg.Seq, nil
}
