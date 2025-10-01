package srpc

import "io"

type ServiceMethod string // e.g. "Service.Method"

type Request struct {
	ServiceMethod ServiceMethod
	Metadata      Metadata
	Body          io.Reader
}

type Response struct {
	ServiceMethod ServiceMethod // echoes that of the Request
	Metadata      Metadata
	Error         error
	Body          io.Reader
}

type Metadata map[string][]string
