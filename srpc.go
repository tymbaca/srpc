package srpc

type RequestHeader struct {
	ServiceMethod string // format: "Service.Method"
	Metadata      map[string]string
}

type ResponseHeader struct {
	ServiceMethod string // echoes that of the Request
	Metadata      map[string]string
}
