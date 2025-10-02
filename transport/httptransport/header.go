package httptransport

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/tymbaca/srpc"
)

const (
	_serviceMethodKey = "Srpc-Service-Method"
	_metadataKey      = "Srpc-Metadata"
	_errorKey         = "Srpc-Error"
)

func setError(h http.Header) {
	h.Set(_errorKey, "true")
}

func hasError(h http.Header) bool {
	return h.Get(_errorKey) == "true"
}

func toHeader(serviceMethod srpc.ServiceMethod, metadata srpc.Metadata) (http.Header, error) {
	header := make(http.Header)
	header.Set(_serviceMethodKey, string(serviceMethod))

	mdJson, err := json.Marshal(metadata)
	if err != nil {
		return nil, fmt.Errorf("marshal metadata json: %w", err)
	}

	header.Set(_metadataKey, url.QueryEscape(string(mdJson)))

	return header, nil
}

func fromHeader(header http.Header) (srpc.ServiceMethod, srpc.Metadata, error) {
	serviceMethod := srpc.ServiceMethod(header.Get(_serviceMethodKey))

	mdJson, err := url.QueryUnescape(header.Get(_metadataKey))
	if err != nil {
		return "", nil, fmt.Errorf("query unescape metadata: %w", err)
	}

	md := make(srpc.Metadata)
	err = json.Unmarshal([]byte(mdJson), &md)
	if err != nil {
		return "", nil, fmt.Errorf("unmarshal metadata json: %w", err)
	}

	return serviceMethod, md, nil
}
