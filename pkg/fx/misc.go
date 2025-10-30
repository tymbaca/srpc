package fx

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
