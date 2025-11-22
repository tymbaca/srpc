package chunked

import (
	"io"
	"math"
	"slices"
)

// NewWriter create a new [Writer].
func NewWriter(w io.Writer) *Writer {
	return &Writer{w: w}
}

// Writer writes chunks into the backing w. It must be closed after all
// data is sent, in order to signal [io.EOF] on the reading end.
type Writer struct {
	w io.Writer
}

// Write writes p in the chunk. It does nothing if len(p) == 0.
func (w *Writer) Write(p []byte) (n int, err error) {
	// Empty chunk is interpreted as EOF on the reader side
	// To send final chunk caller must invoke Close
	if len(p) == 0 {
		return 0, nil
	}

	if len(p) < math.MaxUint16 {
		return w.write(p)
	}

	var nn int
	for pp := range slices.Chunk(p, math.MaxUint16) {
		nn, err = w.write(pp)
		n += nn
		if err != nil {
			break
		}
	}

	return n, err
}

// Close sends a chunk with zero length, signaling reading side the [io.EOF].
func (w *Writer) Close() error {
	_, err := w.write(nil)
	return err
}

// write(nil) will send empty chunk signaling [io.EOF], see [Writer.Close].
func (w *Writer) write(p []byte) (n int, err error) {
	chunk := chunk{Len: uint16(len(p)), Data: p}
	return writeChunk(w.w, chunk)
}
