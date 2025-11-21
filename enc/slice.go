package enc

import (
	"fmt"

	"github.com/tymbaca/srpc/internal/fx"
)

// Slice holds a sequence of items.
type Slice[T any] struct {
	Len  uint32 `sbin:"lenof:Data"`
	Data []T
}

// NewSlice creates new [Slice] from values.
func NewSlice[T any](vs ...T) Slice[T] {
	return Slice[T]{uint32(len(vs)), vs}
}

func (s Slice[T]) String() string {
	return fmt.Sprintf("%v", s.Data)
}

func stringSlice(ss Slice[String]) []string {
	return fx.Map(ss.Data, func(s String) string { return s.Data })
}
