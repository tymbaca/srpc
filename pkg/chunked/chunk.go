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

	return w.Write(chunk.Data)
}
