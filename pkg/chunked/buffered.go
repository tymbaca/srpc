package chunked

import (
	"bufio"
	"io"
)

const defaultBufSize = 4096

func NewBufferWriter(w io.Writer) *BufferedWriter {
	return NewBufferWriterSize(w, defaultBufSize)
}

func NewBufferWriterSize(w io.Writer, size int) *BufferedWriter {
	chunkW := NewWriter(w)
	return &BufferedWriter{
		bufw: bufio.NewWriterSize(chunkW, size),
		w:    chunkW,
	}
}

type BufferedWriter struct {
	bufw *bufio.Writer
	w    *Writer
}

func (w *BufferedWriter) Write(p []byte) (n int, err error) {
	return w.bufw.Write(p)
}

// Close flushed buffered data in chunk and then sends another chunk with zero
// length, signaling reading side the [io.EOF].
func (w *BufferedWriter) Close() error {
	if err := w.bufw.Flush(); err != nil {
		return err
	}

	return w.w.Close()
}
