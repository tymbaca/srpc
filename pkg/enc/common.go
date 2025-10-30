package enc

import (
	"strings"

	"github.com/tymbaca/srpc/pkg/fx"
)

type ServiceMethod String // e.g. "Service.Method"

func (sm ServiceMethod) Split() (service string, method string, ok bool) {
	parts := strings.Split(sm.Data, ".")
	if len(parts) != 2 {
		return "", "", false
	}

	return parts[0], parts[1], true
}

type Metadata Slice[MetadataPair]

func NewMetadata(m map[string][]string) Metadata {
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
