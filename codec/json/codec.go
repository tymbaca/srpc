package json

import (
	"encoding/json"

	"github.com/tymbaca/srpc/codec/helper"
)

var Codec = helper.ToCodec(json.NewEncoder, json.NewDecoder)
