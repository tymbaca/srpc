package codechelp

import (
	"github.com/tymbaca/srpc"
)

func ToCodec[E commonEncoder, D commonDecoder](
	newEncFunc NewEncoderFunc[E],
	newDecFunc NewDecoderFunc[D],
) srpc.Codec {
	return &commonCodecWrapper[E, D]{ToEncoder(newEncFunc), ToDecoder(newDecFunc)}
}

type commonCodecWrapper[E commonEncoder, D commonDecoder] struct {
	srpc.Encoder
	srpc.Decoder
}
