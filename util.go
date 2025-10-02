package srpc

func assert(cond bool) {
	if !cond {
		panic("assertion failure")
	}
}

func tern[T any](cond bool, a, b T) T {
	if cond {
		return a
	} else {
		return b
	}
}
