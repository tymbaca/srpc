package srpc

import (
	"io"
	"strings"
)

type ServiceMethod string // e.g. "Service.Method"

func (sm ServiceMethod) Split() (service string, method string, ok bool) {
	parts := strings.Split(string(sm), ".")
	if len(parts) != 2 {
		return "", "", false
	}

	return parts[0], parts[1], true
}

type Request struct {
	ServiceMethod ServiceMethod
	Metadata      Metadata
	Body          io.Reader // TODO: close?
}

type Response struct {
	ServiceMethod ServiceMethod // echoes that of the Request
	Metadata      Metadata
	StatusCode    StatusCode
	Error         error
	Body          io.Reader // TODO: close?
}

type Metadata map[string][]string

type StatusCode int

func (s StatusCode) String() string {
	switch s {
	case StatusOK:
		return "StatusOK"
	case StatusErrorFromService:
		return "StatusErrorFromService"
	case StatusInvalidServiceMethod:
		return "StatusInvalidServiceMethod"
	case StatusBadRequest:
		return "StatusBadRequest"
	case StatusServiceNotFound:
		return "StatusServiceNotFound"
	case StatusMethodNotFound:
		return "StatusMethodNotFound"
	case StatusInternalError:
		return "StatusInternalError"
	}

	return ""
}

// TODO: remove iota
const (
	StatusOK StatusCode = iota
	StatusErrorFromService
	StatusInvalidServiceMethod
	StatusServiceNotFound
	StatusMethodNotFound
	StatusBadRequest
	StatusInternalError
)
