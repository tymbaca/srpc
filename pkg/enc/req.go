package enc

import (
	"encoding/binary"
	"fmt"
	"io"

	"github.com/tymbaca/sbinary"
)

type Request struct {
	ServiceMethod ServiceMethod
	Metadata      Metadata
	Body          io.Reader `sbin:"-"`
}

func ReadRequest(r io.Reader) (Request, error) {
	var req Request
	if err := sbinary.NewDecoder(r).Decode(&req, binary.BigEndian); err != nil {
		return Request{}, fmt.Errorf("decode request header: %w", err)
	}

	req.Body = r
	return req, nil
}

func WriteRequest(w io.Writer, req Request) error {
	if err := sbinary.NewEncoder(w).Encode(req, binary.BigEndian); err != nil {
		return fmt.Errorf("encode request header: %w", err)
	}

	if req.Body != nil {
		if _, err := io.Copy(w, req.Body); err != nil {
			return fmt.Errorf("write request body: %w", err)
		}
	}

	return nil
}
