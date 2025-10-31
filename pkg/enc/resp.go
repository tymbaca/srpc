package enc

import (
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

func (e *Codec) ReadResponse(r io.Reader) (Response, error) {
	var resp Response

	ver, err := e.checkVersion(r)
	if err != nil {
		return Response{}, err
	}
	resp.Version = ver

	if err := sbinary.NewDecoder(r).Decode(&resp, binary.BigEndian); err != nil {
		return Response{}, fmt.Errorf("decode response header: %w", err)
	}

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

func (e *Codec) WriteResponse(w io.Writer, resp Response) error {
	resp.Version = e.Version

	if err := writeVersion(w, resp.Version); err != nil {
		return err
	}

	if err := sbinary.NewEncoder(w).Encode(resp, binary.BigEndian); err != nil {
		return fmt.Errorf("encode response header: %w", err)
	}

	if resp.Error != nil {
		if _, err := w.Write([]byte(resp.Error.Error())); err != nil {
			return fmt.Errorf("write response error: %w", err)
		}
	}

	if resp.Body != nil {
		if _, err := io.Copy(w, resp.Body); err != nil {
			return fmt.Errorf("write response body: %w", err)
		}
	}

	return nil
}
