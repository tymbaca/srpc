package codechelp

import (
	"io"

	"github.com/tymbaca/srpc"
)

func ToDecoder[T commonDecoder](newFunc NewDecoderFunc[T]) srpc.Decoder {
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
