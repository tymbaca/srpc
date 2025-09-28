package pipe

import (
	"encoding/json"
	"io"
)

func Encode() {
	r, w := io.Pipe()

	e := json.NewEncoder(w)
	e.SetIndent(enc.prefix, enc.indent)
	e.SetEscapeHTML(enc.escapeHTML)

	go func() {
		err := e.Encode(enc.v)
		if err != nil {
			w.CloseWithError(err)
			return
		}
		w.Close()
	}()

	return r
}

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
