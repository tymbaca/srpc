package srpc

import (
	"errors"

	"github.com/tymbaca/srpc/pkg/enc"
)

type Status = enc.StatusCode

func StatusFromError(err error) Status {
	if err == nil {
		return enc.StatusOK
	}

	var se statusError
	if errors.As(err, &se) {
		return se.status
	}

	// error returned by [Client.Call] MUST have a statusError,
	// so this is unreachable
	return enc.StatusInternalError
}

type statusError struct {
	status Status
}

func (e statusError) Error() string {
	return e.status.String()
}
