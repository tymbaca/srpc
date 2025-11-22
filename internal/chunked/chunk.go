// Package chunked provides chunk-based IO,
// inspired by Chunked Transfer Coding (RFC 9112 §7.1).
package chunked

import (
	"encoding/binary"
	"io"
)

type chunk struct {
	Len  uint16
	Data []byte
}

func readChunk(r io.Reader, chunk *chunk) (n int, err error) {
	if err := binary.Read(r, binary.BigEndian, &chunk.Len); err != nil {
		return 0, err
	}

	chunk.Data = make([]byte, chunk.Len)
	return io.ReadFull(r, chunk.Data)
}

func writeChunk(w io.Writer, chunk chunk) (n int, err error) {
	if err := binary.Write(w, binary.BigEndian, chunk.Len); err != nil {
		return 0, err
	}

	if chunk.Len == 0 {
		return 0, nil
	}

	return w.Write(chunk.Data)
}
