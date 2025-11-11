package call

import (
	"github.com/tymbaca/srpc/metadata"
	"github.com/tymbaca/srpc/pkg/enc"
	"github.com/tymbaca/srpc/transport"
)

type Suite struct {
	Req          enc.Request
	Conn         transport.Conn
	RespMetadata *metadata.Metadata
}
