package helper

import (
	"io"

	"github.com/tymbaca/srpc/codec"
)

// ToEncoder creates [codec.Encoder] from common encoder creation function, e.g. `json.NewEncoder`.
func ToEncoder[T commonEncoder](newFunc NewEncoderFunc[T]) codec.Encoder {
	return &commonEncoderWrapper[T]{newFunc: newFunc}
}

type commonEncoderWrapper[T commonEncoder] struct {
	newFunc NewEncoderFunc[T]
}

type NewEncoderFunc[T commonEncoder] = func(w io.Writer) T

type commonEncoder interface {
	Encode(any) error
}

func (co *commonEncoderWrapper[T]) Encode(w io.Writer, src any) error {
	return co.newFunc(w).Encode(src)
}
