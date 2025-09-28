package srpc

import (
	"encoding/json"
	"encoding/xml"
	"testing"

	"github.com/tymbaca/srpc/codechelp"
)

func TestRPC(t *testing.T) {
	jsonDec := codechelp.ToDecoder(json.NewDecoder)
	xmlDec := codechelp.ToDecoder(xml.NewDecoder)

	_ = jsonDec
	_ = xmlDec
}
