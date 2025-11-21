package codec

import "io"

// Codec wraps [Encoder] and [Decoder].
type Codec interface {
	Encoder
	Decoder
}

// Decoder decodes dst from r. Dst must be non-nil pointer.
type Decoder interface {
	Decode(r io.Reader, dst any) error
}

// Encoder encodes src into w.
type Encoder interface {
	Encode(w io.Writer, src any) error
}
