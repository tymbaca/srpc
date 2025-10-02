package srpc

import "io"

type Codec interface {
	Encoder
	Decoder
}

// TODO: way to distinguish io.Reader error and parsing error (e.g. when request
// doesn't match the target type)

type Decoder interface {
	Decode(r io.Reader, dst any) error
}

type Encoder interface {
	Encode(w io.Writer, src any) error
}
