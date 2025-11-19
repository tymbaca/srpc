package chunked

import (
	"bufio"
	"io"
)

const defaultBufSize = 4096

// NewBufferWriter is the same as [NewBufferWriterSize] but it uses
// the default buffer size.
func NewBufferWriter(w io.Writer) *BufferedWriter {
	return NewBufferWriterSize(w, defaultBufSize)
}

// NewBufferWriterSize craetes new [BufferedWriter].
func NewBufferWriterSize(w io.Writer, size int) *BufferedWriter {
	chunkW := NewWriter(w)
	return &BufferedWriter{
		bufw: bufio.NewWriterSize(chunkW, size),
		w:    chunkW,
	}
}

// BufferedWriter wraps base [Writer] with buffer. It must be closed after all
// writes are done.
type BufferedWriter struct {
	bufw *bufio.Writer
	w    *Writer
}

// Write implements [io.Writer].
func (w *BufferedWriter) Write(p []byte) (n int, err error) {
	return w.bufw.Write(p)
}

// Flush sends buffered bytes in chunks(-s).
func (w *BufferedWriter) Flush() error {
	return w.bufw.Flush()
}

// Close flushes buffered data in chunk(-s) and then sends another chunk with zero
// length, signaling [io.EOF] to the reading side.
func (w *BufferedWriter) Close() error {
	if err := w.bufw.Flush(); err != nil {
		return err
	}

	return w.w.Close()
}
