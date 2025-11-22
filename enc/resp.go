package enc

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"

	"github.com/tymbaca/sbinary"
	"github.com/tymbaca/srpc/internal/chunked"
	"github.com/tymbaca/srpc/internal/fx"
	"github.com/tymbaca/srpc/status"
)

// Response is a response header.
type Response struct {
	Version    Version `sbin:"-"` // set by [Encoder]
	StatusCode StatusCode
	Metadata   Metadata
	Error      error     `sbin:"-"`
	Body       io.Reader `sbin:"-"` // nil if Error != nil
}

// ReadResponse reads response from r
func ReadResponse(c Context, r io.Reader) (Response, error) {
	cr := chunked.NewReader(r)
	var resp Response

	ver, err := checkVersion(c, cr)
	if err != nil {
		return Response{}, err
	}
	resp.Version = ver

	if err := sbinary.NewDecoder(cr).Decode(&resp, binary.BigEndian); err != nil {
		return Response{}, fmt.Errorf("decode response header: %w", err)
	}

	if resp.StatusCode == status.OK {
		resp.Body = cr
	} else {
		errBytes, err := io.ReadAll(cr)
		if err != nil {
			return Response{}, fmt.Errorf("read response error: %w", err)
		}

		resp.Error = errors.New(string(errBytes))
	}

	return resp, nil
}

// WriteResponse writes response into w.
// See [WriteRequest].
func WriteResponse(c Context, w io.Writer, resp Response) (err error) {
	defer fx.CloseIfCloser(resp.Body) // if body is pipe.ToReader we need to kill it's goroutine, in case if error happens and we exit before EOF

	resp.Version = c.Version
	cw := chunked.NewBufferWriter(w)
	defer func() {
		err = errors.Join(err, cw.Close())
	}()

	if err := writeVersion(cw, resp.Version); err != nil {
		return err
	}

	if err := sbinary.NewEncoder(cw).Encode(resp, binary.BigEndian); err != nil {
		return fmt.Errorf("encode response header: %w", err)
	}

	if resp.Error != nil {
		resp.Body = bytes.NewBuffer([]byte(resp.Error.Error()))
	}
	if resp.Body == nil {
		resp.Body = bytes.NewBuffer(nil)
	}

	if _, err := io.Copy(cw, resp.Body); err != nil {
		return fmt.Errorf("write response body: %w", err)
	}

	return nil
}
