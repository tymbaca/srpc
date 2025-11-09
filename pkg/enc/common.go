package enc

import (
	"strings"

	"github.com/tymbaca/srpc/pkg/fx"
)

type ServiceMethod = String // e.g. "Service.Method"

func (sm ServiceMethod) Split() (service string, method string, ok bool) {
	parts := strings.Split(sm.Data, ".")
	if len(parts) != 2 {
		return "", "", false
	}

	return parts[0], parts[1], true
}

type Metadata Slice[MetadataPair]

func NewMetadata(m map[string][]string) Metadata {
	if len(m) == 0 {
		return Metadata{}
	}

	pairs := make([]MetadataPair, 0, len(m))

	for k, vals := range m {
		pairs = append(pairs, MetadataPair{
			Key:  NewString(k),
			Vals: NewSlice(fx.Map(vals, NewString)...),
		})
	}

	return Metadata(NewSlice(pairs...))
}

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

type StatusCode uint8

const (
	StatusOK                   StatusCode = 0
	StatusErrorFromService     StatusCode = 1 // Error from server application code
	StatusInvalidServiceMethod StatusCode = 2
	StatusServiceNotFound      StatusCode = 3
	StatusMethodNotFound       StatusCode = 4
	StatusInvalidArgument      StatusCode = 5 // Client passed invalid argument, e.g. when failed to encode/decode request body
	StatusInternalError        StatusCode = 6 // Special case: returned by [Client.Call] when failed to decode server response body
)

func (s StatusCode) String() string {
	switch s {
	case StatusOK:
		return "OK"
	case StatusErrorFromService:
		return "ErrorFromService"
	case StatusInvalidServiceMethod:
		return "InvalidServiceMethod"
	case StatusInvalidArgument:
		return "InvalidArgument"
	case StatusServiceNotFound:
		return "ServiceNotFound"
	case StatusMethodNotFound:
		return "MethodNotFound"
	case StatusInternalError:
		return "InternalError"
	}

	return ""
}
