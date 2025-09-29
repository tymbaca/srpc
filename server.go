package srpc

func NewServer(l ServerListener, codec Codec) *Server {
}

func Register[T any](server *Server, impl T) {
}

type Server struct {
	serviceName
}

// type method struct {
// 	name
// }
