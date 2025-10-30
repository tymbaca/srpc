package enc

import (
	"io"
	"strings"
)

type ServiceMethod String // e.g. "Service.Method"

func (sm ServiceMethod) Split() (service string, method string, ok bool) {
	parts := strings.Split(sm.Data, ".")
	if len(parts) != 2 {
		return "", "", false
	}

	return parts[0], parts[1], true
}

type Request struct {
	ServiceMethod ServiceMethod
	Metadata      Metadata
	Body          io.Reader `sbin:"-"`
}

type Response struct {
	ServiceMethod ServiceMethod // echoes that of the Request
	Metadata      Metadata
	StatusCode    StatusCode
	Error         error
	Body          io.Reader `sbin:"-"`
}

type Metadata Slice[MetadataPair]

func (m Metadata) Map() map[string][]string {
	res := make(map[string][]string, m.Len)
	for _, pair := range m.Data {
		res[pair.Key.Data] = stringSlice(pair.Vals)
	}

	return res
}

type MetadataPair struct {
	Key  String
	Vals Slice[String]
}

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

func readReq(r io.Reader) (Request, error) {
	var req Request

}

func writeReq(w io.Writer, req Request) error
func readResp(r io.Reader) (Response, error)
func writeResp(w io.Writer, req Response) error
