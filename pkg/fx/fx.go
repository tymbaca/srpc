// Package fx provides commond utility funcitons
package fx

import (
	"context"
	"io"
)

func Assert(cond bool) {
	if !cond {
		panic("assertion failure")
	}
}

func Tern[T any](cond bool, a, b T) T {
	if cond {
		return a
	} else {
		return b
	}
}

func Map[A, B any](input []A, conv func(a A) B) []B {
	output := make([]B, len(input))
	for i, a := range input {
		output[i] = conv(a)
	}

	return output
}

func CloseOnCancel(ctx context.Context, c io.Closer) (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithCancel(ctx)
	go func() {
		<-ctx.Done()
		c.Close()
	}()

	return ctx, cancel
}

func CloseIfCloser(v any) error {
	if c, ok := v.(io.Closer); ok {
		return c.Close()
	}

	return nil
}
