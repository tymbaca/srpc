package enc

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"

	"github.com/tymbaca/sbinary"
)

type Response struct {
	Version    Version `sbin:"-"` // set by [Encoder]
	StatusCode StatusCode
	Metadata   Metadata
	Error      error     `sbin:"-"`
	Body       io.Reader `sbin:"-"` // nil if Error != nil
}

func ReadResponse(c Context, r io.Reader) (Response, error) {
	var resp Response

	ver, err := checkVersion(c, r)
	if err != nil {
		return Response{}, err
	}
	resp.Version = ver

	dec := sbinary.NewDecoder(r)
	if err := dec.Decode(&resp, binary.BigEndian); err != nil {
		return Response{}, fmt.Errorf("decode response header: %w", err)
	}
	var bh bodyHeader
	if err := dec.Decode(&bh, binary.BigEndian); err != nil {
		return Response{}, fmt.Errorf("decode body header: %w", err)
	}

	r = io.LimitReader(r, int64(bh.Size))

	if resp.StatusCode == StatusOK {
		resp.Body = r
	} else {
		errBytes, err := io.ReadAll(r)
		if err != nil {
			return Response{}, fmt.Errorf("read response error: %w", err)
		}

		resp.Error = errors.New(string(errBytes))
	}

	return resp, nil
}

// WriteResponse writes response into w.
// See [WriteRequest].
func WriteResponse(c Context, w io.Writer, resp Response) error {
	resp.Version = c.Version

	if err := writeVersion(w, resp.Version); err != nil {
		return err
	}

	if err := sbinary.NewEncoder(w).Encode(resp, binary.BigEndian); err != nil {
		return fmt.Errorf("encode response header: %w", err)
	}

	if resp.Error != nil {
		resp.Body = bytes.NewBuffer([]byte(resp.Error.Error()))
	}

	if resp.Body == nil {
		resp.Body = bytes.NewBuffer(nil)
	}

	var bodyLen int
	switch b := resp.Body.(type) {
	case *bytes.Buffer:
		bodyLen = b.Len()
	default:
		return fmt.Errorf("currently resp.Body must be [*bytes.Buffer], got: %#v", b)
	}

	if err := sbinary.NewEncoder(w).Encode(bodyHeader{Size: uint64(bodyLen)}, binary.BigEndian); err != nil {
		return fmt.Errorf("write response body header: %w", err)
	}
	if _, err := io.Copy(w, resp.Body); err != nil {
		return fmt.Errorf("write response body: %w", err)
	}

	return nil
}
