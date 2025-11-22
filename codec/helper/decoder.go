package helper

import (
	"io"

	"github.com/tymbaca/srpc/codec"
)

// ToDecoder creates [codec.Decoder] from common decoder creation function, e.g. `json.NewDecoder`.
func ToDecoder[T commonDecoder](newFunc NewDecoderFunc[T]) codec.Decoder {
	return &commonDecoderWrapper[T]{newFunc: newFunc}
}

type commonDecoderWrapper[T commonDecoder] struct {
	newFunc NewDecoderFunc[T]
}

type NewDecoderFunc[T commonDecoder] = func(r io.Reader) T

type commonDecoder interface {
	Decode(any) error
}

func (co *commonDecoderWrapper[T]) Decode(r io.Reader, dst any) error {
	return co.newFunc(r).Decode(dst)
}
