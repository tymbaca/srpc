package pipe

import (
	"io"
)

func ToReader(fn func(w io.Writer) error) io.Reader {
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
