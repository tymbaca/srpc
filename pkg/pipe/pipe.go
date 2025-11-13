package pipe

import (
	"fmt"
	"io"
)

func ToReader(fn func(w io.Writer) error) io.ReadCloser {
	r, w := io.Pipe()

	go func() {
		fmt.Println("PIPE WRITER CREATE")
		defer fmt.Println("PIPE WRITER EXIT")
		err := fn(w)
		if err != nil {
			fmt.Println("PIPE WRITER ERROR", err)
			w.CloseWithError(err)
			return
		}
		w.Close()
		fmt.Println("PIPE WRITER NO ERROR")
	}()

	return r
}
