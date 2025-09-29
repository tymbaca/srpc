package srpc

func NewClient[T any]() *Client[T] {
	return &Client[T]{}
}

type Client[T any] struct {
	Inner T
}
