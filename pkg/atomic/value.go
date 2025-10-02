package atomic

import "sync/atomic"

type Value[T any] struct {
	val atomic.Value
}

func (v *Value[T]) CompareAndSwap(old T, new T) (swapped bool) {
	return v.val.CompareAndSwap(old, new)
}

func (v *Value[T]) Load() (val T) {
	return v.val.Load().(T)
}

func (v *Value[T]) Store(val T) {
	v.val.Store(val)
}

func (v *Value[T]) Swap(new T) (old T) {
	return v.val.Swap(new).(T)
}
