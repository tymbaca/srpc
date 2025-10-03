package httptransport

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	"github.com/tymbaca/srpc"
)

const (
	_serviceMethodKey = "Srpc-Service-Method"
	_metadataKey      = "Srpc-Metadata"
	_statusKey        = "Srpc-Status"
	_errorKey         = "Srpc-Error"
)

func setStatus(h http.Header, status srpc.StatusCode) {
	h.Set(_statusKey, strconv.Itoa(int(status)))
}

func getStatus(h http.Header) (status srpc.StatusCode, err error) {
	statusStr := h.Get(_statusKey)
	if statusStr == "" {
		return 0, fmt.Errorf("no status code in header")
	}

	statusInt, err := strconv.Atoi(statusStr)
	if err != nil {
		return 0, fmt.Errorf("convert status from header to int: %w", err)
	}

	return srpc.StatusCode(statusInt), nil
}

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
