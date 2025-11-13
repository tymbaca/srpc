package enc

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"

	"github.com/tymbaca/sbinary"
	"github.com/tymbaca/srpc/pkg/chunked"
	"github.com/tymbaca/srpc/pkg/fx"
)

type Request struct {
	Version       Version `sbin:"-"`
	ServiceMethod ServiceMethod
	Metadata      Metadata
	Body          io.Reader `sbin:"-"`
}

func ReadRequest(c Context, r io.Reader) (Request, error) {
	cr := chunked.NewReader(r)
	var req Request

	ver, err := checkVersion(c, cr)
	if err != nil {
		return Request{}, err
	}
	req.Version = ver

	if err := sbinary.NewDecoder(cr).Decode(&req, binary.BigEndian); err != nil {
		return Request{}, fmt.Errorf("decode request header: %w", err)
	}

	req.Body = cr
	return req, nil
}

// WriteRequest writes request into w.
// Currently req.Body must be [*bytes.Buffer] when writing.
// In future, chunked io will be needed for dynamically filled readers.
func WriteRequest(c Context, w io.Writer, req Request) (err error) {
	defer fx.CloseIfCloser(req.Body) // if body is pipe.ToReader we need to kill it's goroutine, in case if error happens and we exit before EOF

	req.Version = c.Version
	cw := chunked.NewBufferWriter(w)
	defer func() {
		err = errors.Join(err, cw.Close())
	}()

	if err := writeVersion(cw, req.Version); err != nil {
		return err
	}

	if err := sbinary.NewEncoder(cw).Encode(req, binary.BigEndian); err != nil {
		return fmt.Errorf("encode request header: %w", err)
	}

	if req.Body == nil {
		req.Body = bytes.NewBuffer(nil)
	}

	if _, err := io.Copy(cw, req.Body); err != nil {
		return fmt.Errorf("write request body: %w", err)
	}

	return nil
}
