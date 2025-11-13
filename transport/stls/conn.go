package stls

import (
	"crypto/cipher"

	transport "github.com/tymbaca/srpc/transport"
)

type Conn struct {
	r       cipher.StreamReader
	w       cipher.StreamWriter
	backing transport.Conn
}

func connWithCipher(conn transport.Conn, stream cipher.Stream) *Conn {
	return &Conn{
		r:       cipher.StreamReader{S: stream, R: conn},
		w:       cipher.StreamWriter{S: stream, W: conn},
		backing: conn,
	}
}

func (c *Conn) Read(p []byte) (n int, err error) {
	return c.r.Read(p)
}

func (c *Conn) Write(p []byte) (n int, err error) {
	return c.w.Write(p)
}

func (c *Conn) RemoteAddr() string {
	return c.backing.RemoteAddr()
}

func (c *Conn) LocalAddr() string {
	return c.backing.LocalAddr()
}

func (c *Conn) Close() error {
	return c.backing.Close()
}
