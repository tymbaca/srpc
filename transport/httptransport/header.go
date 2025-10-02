package httptransport

import (
	"net/http"

	"github.com/tymbaca/srpc"
)

func toHeader(serviceMethod srpc.ServiceMethod, metadata srpc.Metadata) http.Header {
	header := make(http.Header, len(metadata)+1)
}

func fromHeader(http.Header) (srpc.ServiceMethod, srpc.Metadata) {
}
