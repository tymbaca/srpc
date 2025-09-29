package codechelp

import (
	"encoding/gob"
	"encoding/json"
	"encoding/xml"
	"testing"
)

func TestTo(t *testing.T) {
	jsonDec := ToDecoder(json.NewDecoder)
	xmlDec := ToDecoder(xml.NewDecoder)
	gobDec := ToDecoder(gob.NewDecoder)
	_ = jsonDec
	_ = xmlDec
	_ = gobDec

	jsonEnc := ToEncoder(json.NewEncoder)
	xmlEnc := ToEncoder(xml.NewEncoder)
	gobEnc := ToEncoder(gob.NewEncoder)
	_ = jsonEnc
	_ = xmlEnc
	_ = gobEnc

	jsonCodec := ToCodec(json.NewEncoder, json.NewDecoder)
	xmlCodec := ToCodec(xml.NewEncoder, xml.NewDecoder)
	gobCodec := ToCodec(gob.NewEncoder, gob.NewDecoder)
	_ = jsonCodec
	_ = xmlCodec
	_ = gobCodec
}
