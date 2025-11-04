// Package chunked provides chunk-based IO,
// inspired by Chunked Transfer Coding (RFC 9112 §7.1).
package chunked

import (
	"bufio"
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

// write(nil) will send empty chunk signaling [io.EOF], see [Writer.Close].
func (w *Writer) write(p []byte) (n int, err error) {
	chunk := chunk{Len: uint16(len(p)), Data: p}
	return writeChunk(w.w, chunk)
}

// Close flushed buffered data in chunk and then sends another chunk with zero
// length, signaling reading side the [io.EOF].
func (w *Writer) Close() error {
	_, err := w.write(nil)
	return err
}

type BufferedWriter struct {
	bufw *bufio.Writer
	w    io.Writer
}

func (wr *BufferedWriter) Write(p []byte) (n int, err error) {
	panic("not implemented") // TODO: Implement
}

// Close flushed buffered data in chunk and then sends another chunk with zero
// length, signaling reading end the [io.EOF].
func (wr *BufferedWriter) Close() error {
	panic("not implemented") // TODO: Implement
}
