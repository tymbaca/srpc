package codec

import "encoding/json"

var JSON = ToCodec(json.NewEncoder, json.NewDecoder)
