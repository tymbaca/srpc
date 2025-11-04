// Package chunked provides chunk-based IO,
// inspired by Chunked Transfer Coding (RFC 9112 §7.1).
package chunked

import (
	"io"
	"math"
	"slices"
)

func NewWriter(w io.Writer) *Writer {
	return &Writer{w: w}
}

type Writer struct {
	w io.Writer
}

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

// Close sends another chunk with zero length, signaling reading side the [io.EOF].
func (w *Writer) Close() error {
	_, err := w.write(nil)
	return err
}

// write(nil) will send empty chunk signaling [io.EOF], see [Writer.Close].
func (w *Writer) write(p []byte) (n int, err error) {
	chunk := chunk{Len: uint16(len(p)), Data: p}
	return writeChunk(w.w, chunk)
}
