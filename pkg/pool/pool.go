package pool

import "sync"

func New[T any](newFunc func() T) *Pool[T] {
	return &Pool[T]{
		pool: sync.Pool{
			New: func() any { return newFunc() },
		},
	}
}

// Pool must be created with [New]
type Pool[T any] struct {
	pool sync.Pool
}

func (p *Pool[T]) Get() T {
	return p.pool.Get().(T)
}

func (p *Pool[T]) Put(x T) {
	p.pool.Put(x)
}
