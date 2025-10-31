package enc

import (
	"encoding/binary"
	"fmt"
	"io"

	"github.com/tymbaca/sbinary"
)

type Request struct {
	Version       Version `sbin:"-"`
	ServiceMethod ServiceMethod
	Metadata      Metadata
	Body          io.Reader `sbin:"-"`
}

func (e *Codec) ReadRequest(r io.Reader) (Request, error) {
	var req Request

	ver, err := e.checkVersion(r)
	if err != nil {
		return Request{}, err
	}
	req.Version = ver

	if err := sbinary.NewDecoder(r).Decode(&req, binary.BigEndian); err != nil {
		return Request{}, fmt.Errorf("decode request header: %w", err)
	}

	req.Body = r
	return req, nil
}

func (e *Codec) WriteRequest(w io.Writer, req Request) error {
	req.Version = e.Version

	if err := writeVersion(w, req.Version); err != nil {
		return err
	}

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
