// Package json wraps standard library `json` package into srpc codec.
package json

import (
	"encoding/json"

	"github.com/tymbaca/srpc/codec/helper"
)

// Codec is a [codec.Codec] that wraps standard library `json` package.
var Codec = helper.ToCodec(json.NewEncoder, json.NewDecoder)
