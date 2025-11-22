package enc

import (
	"strings"

	"github.com/tymbaca/srpc/internal/fx"
	"github.com/tymbaca/srpc/status"
)

// ServiceMethod represents service and method string
type ServiceMethod = String // e.g. "Service.Method"

// Split splits ServiceMethod into service and method strings.
func (sm ServiceMethod) Split() (service string, method string, ok bool) {
	parts := strings.Split(sm.Data, ".")
	if len(parts) != 2 {
		return "", "", false
	}

	return parts[0], parts[1], true
}

// Metadata holds metadata key-values.
type Metadata Slice[MetadataPair]

// NewMetadata creates new [Metadata] from map.
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

// Map returns Go map representation of [Metadata].
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

type StatusCode = status.Code
