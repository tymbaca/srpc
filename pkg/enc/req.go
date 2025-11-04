package enc

import (
	"bytes"
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

func ReadRequest(c Context, r io.Reader) (Request, error) {
	var req Request

	ver, err := checkVersion(c, r)
	if err != nil {
		return Request{}, err
	}
	req.Version = ver

	dec := sbinary.NewDecoder(r)

	if err := dec.Decode(&req, binary.BigEndian); err != nil {
		return Request{}, fmt.Errorf("decode request header: %w", err)
	}
	var bh bodyHeader
	if err := dec.Decode(&bh, binary.BigEndian); err != nil {
		return Request{}, fmt.Errorf("decode body header: %w", err)
	}

	req.Body = io.LimitReader(r, int64(bh.Size))
	return req, nil
}

// WriteRequest writes request into w.
// Currently req.Body must be [*bytes.Buffer] when writing.
// In future, chunked io will be needed for dynamically filled readers.
func WriteRequest(c Context, w io.Writer, req Request) error {
	req.Version = c.Version

	if err := writeVersion(w, req.Version); err != nil {
		return err
	}

	if err := sbinary.NewEncoder(w).Encode(req, binary.BigEndian); err != nil {
		return fmt.Errorf("encode request header: %w", err)
	}

	if req.Body == nil {
		req.Body = bytes.NewBuffer(nil)
	}

	var bodyLen int
	switch b := req.Body.(type) {
	case *bytes.Buffer:
		bodyLen = b.Len()
	default:
		return fmt.Errorf("currently req.Body must be [*bytes.Buffer], got: %#v", b)
	}

	if err := sbinary.NewEncoder(w).Encode(bodyHeader{Size: uint64(bodyLen)}, binary.BigEndian); err != nil {
		return fmt.Errorf("write request body header: %w", err)
	}
	if _, err := io.Copy(w, req.Body); err != nil {
		return fmt.Errorf("write request body: %w", err)
	}

	return nil
}

type bodyHeader struct {
	Size uint64
}
