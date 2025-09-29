package srpc

import "io"

type Codec interface {
	Encoder
	Decoder
}

type Decoder interface {
	Decode(r io.Reader, dst any) error
}

type Encoder interface {
	Encode(w io.Writer, src any) error
}
