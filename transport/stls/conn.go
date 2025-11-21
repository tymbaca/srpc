package stls

import (
	"crypto/cipher"

	transport "github.com/tymbaca/srpc/transport"
)

func connWithCipher(c transport.Conn, stream cipher.Stream) *conn {
	return &conn{
		r:       cipher.StreamReader{S: stream, R: c},
		w:       cipher.StreamWriter{S: stream, W: c},
		backing: c,
	}
}

type conn struct {
	r       cipher.StreamReader
	w       cipher.StreamWriter
	backing transport.Conn
}

func (c *conn) Read(p []byte) (n int, err error) {
	return c.r.Read(p)
}

func (c *conn) Write(p []byte) (n int, err error) {
	return c.w.Write(p)
}

func (c *conn) RemoteAddr() string {
	return c.backing.RemoteAddr()
}

func (c *conn) LocalAddr() string {
	return c.backing.LocalAddr()
}

func (c *conn) Close() error {
	return c.backing.Close()
}
