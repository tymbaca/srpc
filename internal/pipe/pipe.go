// Package pipe provides streaming IO utility.
package pipe

import (
	"io"
)

// ToReader launches provided writing function in the goroutine and returnes
// reading side of the pipe. Returned r must be closed when it no longer needed
// to prevent leaking goroutine in case if writing function still in progress.
func ToReader(fn func(w io.Writer) error) (r io.ReadCloser) {
	r, w := io.Pipe()

	go func() {
		err := fn(w)
		if err != nil {
			w.CloseWithError(err)
			return
		}
		w.Close()
	}()

	return r
}
