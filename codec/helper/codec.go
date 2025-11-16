package helper

import "github.com/tymbaca/srpc/codec"

func ToCodec[E commonEncoder, D commonDecoder](
	newEncFunc NewEncoderFunc[E],
	newDecFunc NewDecoderFunc[D],
) codec.Codec {
	return &commonCodecWrapper[E, D]{ToEncoder(newEncFunc), ToDecoder(newDecFunc)}
}

type commonCodecWrapper[E commonEncoder, D commonDecoder] struct {
	codec.Encoder
	codec.Decoder
}
