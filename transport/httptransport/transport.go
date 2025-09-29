package httptransport

import (
	"io"

	"github.com/tymbaca/srpc"
)

type ClientConn struct{}

func NewClientTransport(url string, method string) ClientConn {
	return ClientConn{}
}

func (ht *ClientConn) WriteRequest(header srpc.Request, body io.Reader) error {
	panic("not implemented") // TODO: Implement
}

func (ht *ClientConn) ResponseHeader() (srpc.Response, error) {
	panic("not implemented") // TODO: Implement
}

func (ht *ClientConn) ResponseBody() (io.Reader, error) {
	panic("not implemented") // TODO: Implement
}

func (ht *ClientConn) Close() error {
	panic("not implemented") // TODO: Implement
}
