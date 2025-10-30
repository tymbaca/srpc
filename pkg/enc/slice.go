package enc

import "github.com/tymbaca/srpc/pkg/fx"

type Slice[T any] struct {
	Len  uint32 `sbin:"lenof:Data"`
	Data []T
}

func NewSlice[T any](vs ...T) Slice[T] {
	return Slice[T]{uint32(len(vs)), vs}
}

func stringSlice(ss Slice[String]) []string {
	return fx.Map(ss.Data, func(s String) string { return s.Data })
}
