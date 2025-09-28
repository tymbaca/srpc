package httptransport

import (
	"io"
	"net/http"

	"github.com/tymbaca/srpc"
)

type ClientConn struct{}

func NewClientTransport(url string, method string) ClientConn {
	req := http.NewRequestWithContext()
	http.DefaultClient.Do()
	return ClientConn{}
}

func (ht *ClientConn) WriteRequest(header srpc.RequestHeader, body io.Reader) error {
	panic("not implemented") // TODO: Implement
}

func (ht *ClientConn) ResponseHeader() (srpc.ResponseHeader, error) {
	panic("not implemented") // TODO: Implement
}

func (ht *ClientConn) ResponseBody() (io.Reader, error) {
	panic("not implemented") // TODO: Implement
}

func (ht *ClientConn) Close() error {
	panic("not implemented") // TODO: Implement
}
