package srpc

import "io"

type Decoder interface {
	Decode(r io.Reader, dst any) error
}

type Encoder interface {
	Encode(w io.Writer, src any) error
}
