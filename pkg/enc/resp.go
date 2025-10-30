package enc

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"

	"github.com/tymbaca/sbinary"
)

type Response struct {
	StatusCode StatusCode
	Metadata   Metadata
	Error      error     `sbin:"-"`
	Body       io.Reader `sbin:"-"` // nil if Error != nil
}

func ReadResponse(r io.Reader) (Response, error) {
	var resp Response
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

func WriteResponse(w io.Writer, resp Response) error {
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
