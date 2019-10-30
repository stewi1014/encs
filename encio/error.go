package encio

import (
	"errors"
	"runtime"
)

// Error handling in encs is designed to provide an easy way to distinguish io errors and bad data from internal encoding errors,
// and to reuse a small set of common error kinds for as many errors as possible, with extra information wrapped as applicable.
// Panics are only used when there is a clear misuse of the library; programmer error.
// To this end, I have grouped all error cases into two error wrappers; IOError and Error, the idea being that
// IOError errors indicate a bad io.Reader/io.Writer, and the caller should stop using it, and
// Error errors indicate a caller should stop using an Encodable, or use it in a different way.
//
// In this way, errors can be checked with
//
//	var encErr Error
//	var ioErr IOError
//	if errors.As(err, encError) {
//		//handle encoding error
//	} else if errors.As(err, ioErr) {
//		//handle io error
//	}
//
// These errors will be wrapped by IOError or Error.
var (
	// ErrMalformed is returned when the read data is impossible to decode.
	ErrMalformed = errors.New("malformed")

	// ErrBadType is returned when a type, where possible to detect, is wrong, unresolvable or inappropriate.
	// Due to the usage of unsafe.Pointer, it is not usually possible to detect incorrect types.
	// If this error is seen, it should be taken seriously; encoding of incorrect types has undefined behaviour.
	ErrBadType = errors.New("bad type")

	// ErrNilPointer is returned if a pointer that should not be nil is nil.
	ErrNilPointer = errors.New("nil pointer")

	// ErrBadConfig is returned when the config cannot be used to encode the given encodable.
	// i.e. Config.Resolver = nil when creating Interface Encodables.
	ErrBadConfig = errors.New("bad config")
)

// NewIOError returns an IOError wrapping err with the given message.
// err is typically the error returned from the io.Reader/io.Writer, or another error describing why the reader isn't operating correctly.
// message has extra information about the error; if empty, it is filled with the calling fucntions name.
func NewIOError(err error, message string) error {
	if err == nil {
		return NewError(errors.New("unknown error"), "trying to create new IOError", "io.NewIOError")
	}
	if message == "" {
		message = "in " + GetCaller(1)
	}

	return IOError{
		Err:     err,
		Message: message,
	}
}

// IOError is returned when io errors occour, or when read data is malformed.
type IOError struct {
	Err     error
	Message string
}

// Error implements error
func (e IOError) Error() string {
	if e.Message != "" {
		return e.Message + ": " + e.Err.Error()
	}
	return e.Err.Error()
}

// Unwrap implements errors's Unwrap()
func (e IOError) Unwrap() error {
	return e.Err
}

// NewError returns an Error wrapping err with message and caller.
// If caller is empty, it is automatically filled with the calling functions name.
func NewError(err error, message string, caller string) error {
	if caller == "" {
		caller = GetCaller(1)
	}

	return Error{
		Err: err,
	}
}

// Error is returned when an internal error is encountered while encoding.
type Error struct {
	Err     error
	Message string
	Caller  string
}

// Error implements error
func (e Error) Error() (str string) {
	if e.Caller != "" {
		str = e.Caller + ": "
	}

	str += e.Err.Error()

	if e.Message != "" {
		str += " (" + e.Message + ")"
	}

	return str
}

// Unwrap implements errors's Unwrap()
func (e Error) Unwrap() error {
	return e.Err
}

// GetCaller returns the name of the calling function, skipping skip functions.
// i.e. 0 writes the calling function, 1 the function calling that etc...
func GetCaller(skip int) string {
	pcs := make([]uintptr, 1)
	n := runtime.Callers(2+skip, pcs)
	if n != 1 {
		return "Unknown Function"
	}

	frames := runtime.CallersFrames(pcs)
	frame, _ := frames.Next()
	return frame.Function
}
