package chunked

import (
	"bytes"
	"io"
)

// NewReader creates a new [Reader]
func NewReader(r io.Reader) *Reader {
	buf := bytes.NewBuffer(nil)
	buf.Grow(defaultBufSize)
	return &Reader{r: r, buf: buf}
}

// Reader reads chunks from backing r. When got empty chunk it will return
// [io.EOF] on the next Read call.
type Reader struct {
	r      io.Reader
	buf    *bytes.Buffer
	sawEOF bool
}

// Read implements [io.Reader].
// It reads next chunk for r and copies bytes into p. If p is smaller then
// received chunk, it will reserve remaining bytes into buffer. On the next
// call, if buffer is not empty, it's bytes will be copied into p, without
// reading new chunk from r.
//
// It is not safe for concurrent use.
func (r *Reader) Read(p []byte) (n int, err error) {
	if r.sawEOF {
		return 0, io.EOF
	}

	if r.buf.Len() > 0 {
		return r.buf.Read(p)
	} else {
		// to prevent infinitely growing buffer, when bug chunks are often
		r.buf.Reset()
	}

	var chunk chunk
	_, err = readChunk(r.r, &chunk)
	if chunk.Len == 0 && err == nil {
		r.sawEOF = true
		return 0, io.EOF
	}
	if chunk.Len > 0 {
		n = copy(p, chunk.Data)

		if len(chunk.Data) > len(p) {
			_, _ = r.buf.Write(chunk.Data[len(p):])
		}
	}
	return n, err
}
