package srpc

import "io"

type Request struct {
	ServiceMethod string // format: "Service.Method"
	Metadata      map[string]string
	Body          io.Reader
}

type Response struct {
	ServiceMethod string // echoes that of the Request
	Metadata      map[string]string
	Body          io.Reader
}
