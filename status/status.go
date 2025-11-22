// Package status provides sRPC status code type and helpers for error handling.
package status

import (
	"errors"
	"fmt"
)

// Code represents the status code in the response
type Code uint8

const (
	// OK is returned when server successfully handled the request.
	OK Code = 0
	// ErrorFromService is returned when service code returned error
	// to the server without specifying the status (with [Error] or [Errorf]).
	ErrorFromService Code = 1
	// InvalidServiceMethod is returned when ServiceMethod string in the
	// request is incorrect (doesn't have `Service.Method` format).
	InvalidServiceMethod Code = 2
	// ServiceNotFound is returned when specified service name not found.
	ServiceNotFound Code = 3
	// MethodNotFound is returned when specified method name not found in
	// the service.
	MethodNotFound Code = 4
	// InvalidArgument is returned when client passed invalid argument,
	// e.g. when failed to encode/decode request body.
	InvalidArgument Code = 5
	// InternalError is used for internal srpc errors. Currecly it only
	// returned by [Client.Call] when it failed to decode server response
	// body (i.e. server sent response of invalid format).
	InternalError Code = 6
	// TransportError is used when something went wrong on the transport
	// layer. Can be used by transport implementation.
	TransportError Code = 7
)

// String returns string representation of [Code].
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

// Error creates new error with provided status code and error message.
func Error(status Code, errMsg string) error {
	return statusError{status: status, e: errors.New(errMsg)}
}

// Errorf is the same as [Error] but it formats the errors message with provided args.
func Errorf(status Code, errFmt string, args ...any) error {
	return statusError{status: status, e: fmt.Errorf(errFmt, args...)}
}

// ErrorOnlyCode is the same as [Error] but without additional message.
func ErrorOnlyCode(status Code) error {
	return statusError{status: status, e: nil}
}

// FromError extracts status code from the error.
// It returns false if there was no status code.
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
	if e.e == nil {
		return e.status.String()
	}

	return fmt.Sprintf("%s: %s", e.status.String(), e.e.Error())
}

func (e statusError) Unwrap() error {
	return e.e
}
