package enc

import (
	"fmt"

	"github.com/tymbaca/srpc/pkg/fx"
)

type Slice[T any] struct {
	Len  uint32 `sbin:"lenof:Data"`
	Data []T
}

func NewSlice[T any](vs ...T) Slice[T] {
	return Slice[T]{uint32(len(vs)), vs}
}

func (s Slice[T]) String() string {
	return fmt.Sprintf("%#v", s.Data)
}

func stringSlice(ss Slice[String]) []string {
	return fx.Map(ss.Data, func(s String) string { return s.Data })
}
