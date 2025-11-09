package status

import (
	"errors"
	"fmt"
)

type Code uint8

const (
	OK                   Code = 0
	ErrorFromService     Code = 1 // Set when server application code returned error without specifying the status (with [Error] or [Errorf])
	InvalidServiceMethod Code = 2
	ServiceNotFound      Code = 3
	MethodNotFound       Code = 4
	InvalidArgument      Code = 5 // Client passed invalid argument, e.g. when failed to encode/decode request body
	InternalError        Code = 6 // Special case: returned by [Client.Call] when failed to decode server response body
)

func (s Code) String() string {
	switch s {
	case OK:
		return "OK"
	case ErrorFromService:
		return "ErrorFromService"
	case InvalidServiceMethod:
		return "InvalidServiceMethod"
	case InvalidArgument:
		return "InvalidArgument"
	case ServiceNotFound:
		return "ServiceNotFound"
	case MethodNotFound:
		return "MethodNotFound"
	case InternalError:
		return "InternalError"
	}

	return ""
}

func ErrorOnlyCode(status Code) error {
	return statusError{status: status, e: nil}
}

func Error(status Code, errMsg string) error {
	return statusError{status: status, e: errors.New(errMsg)}
}

func Errorf(status Code, errFmt string, args ...any) error {
	return statusError{status: status, e: fmt.Errorf(errFmt, args...)}
}

func FromError(err error) (Code, bool) {
	var se statusError
	ok := errors.As(err, &se)
	return se.status, ok
}

type statusError struct {
	status Code
	e      error
}

func (e statusError) Error() string {
	return fmt.Sprintf("%s: %s", e.status.String(), e.e.Error())
}

func (e statusError) Unwrap() error {
	return e.e
}
