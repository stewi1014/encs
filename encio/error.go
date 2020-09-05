package encio

import (
	"errors"
	"fmt"
	"reflect"
	"runtime"
)

// Error handling in encs is designed to provide an easy way to distinguish io errors and bad data from internal encoding errors,
// and to reuse a small set of common error kinds for as many errors as possible, with extra information wrapped as applicable.
// To this end, all error cases are grouped into two error wrappers; IOError and Error, the idea being that
// IOError indicates a bad io.Reader/io.Writer, and the caller should stop using it, while
// Error indicates a caller should stop using an Encodable or use it in a different way.
//
// In this way, errors can be checked with
// ```
// var encErr Error
// var ioErr IOError
// if errors.As(err, encError) {
// 	// handle encoding error
// } else if errors.As(err, ioErr) {
//	// handle io error
// }
// ```
//
// Panics are only used when there is a clear misuse of the library; programmer error.
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

	// ErrHashColission is returned when two hashes collide.
	// If this is returned (or panic'd), investigation into encs is required;
	// this should never occur, and is here for completeness.
	ErrHashColission = errors.New("hash colission")
)

// NewIOError returns an IOError wrapping err with the given message.
// err is typically the error returned from the io.Reader/io.Writer, or another error describing why the io.Reader/io.Writer isn't operating correctly.
// Message has extra information about the error.
func NewIOError(err error, device interface{}, message string, depth int) error {
	if _, ok := err.(IOError); ok {
		return err
	}

	if err == nil {
		return NewError(errors.New("unknown error"), "refusing to create IOError with nil error", 0)
	}

	location := GetCaller(depth + 1)

	return IOError{
		Err:      err,
		Device:   device,
		Message:  message,
		Location: location,
	}
}

// IOError is returned for errors external to encs and pertaining to data IO,
// such as corrupted data, unexpected EOFs or bad io.Reader/io.Writer implementations.
//
// IOError implements Unwrap(), so errors.Is can be used; e.g. errors.Is(err, io.ErrClosedPipe) if writing to a pipe.
type IOError struct {
	// Error is the received error.
	Err error

	// Device is the io.Reader or io.Writer that was involved.
	Device interface{}

	// Message contains extra information about the error.
	Message string

	// Location is the name of the function where the error occoured.
	Location string
}

// Error implements error.
func (e IOError) Error() string {
	str := fmt.Sprintf("\"%v\"", e.Err.Error())

	if e.Device != nil {
		str += fmt.Sprintf(" using %T", e.Device)
	}

	if e.Location != "" {
		str += fmt.Sprintf(" in %v", e.Location)
	}

	if e.Message != "" {
		str += fmt.Sprintf(" (%v)", e.Message)
	}

	return str
}

// Unwrap implements errors' Unwrap().
func (e IOError) Unwrap() error {
	return e.Err
}

// NewError returns an Error wrapping err with message and caller.
// Depth is how deep the stack is after the logical location of the error; which function to blame.
// i.e. 0 will use the calling function of NewError, 1 the calling function of that etc...
func NewError(err error, message string, depth int) error {
	if err == nil {
		return NewError(errors.New("unknown error"), "refusing to create Error with nil error", 1)
	}

	caller := GetCaller(depth + 1)

	return Error{
		Err:      err,
		Message:  message,
		Location: caller,
	}
}

// Error is returned for errors originating from the usage of encs.
//
// Error implements Unwrap(), so errors.Is can be used; e.g. errors.Is(err, encio.ErrBadType) to check if an unregistered type was received.
type Error struct {
	Err      error
	Message  string
	Location string
}

// Error implements error.
func (e Error) Error() string {
	str := fmt.Sprintf("\"%v\"", e.Err.Error())

	if e.Location != "" {
		str += fmt.Sprintf(" in %v", e.Location)
	}

	if e.Message != "" {
		str += fmt.Sprintf(" (%v)", e.Message)
	}

	return str
}

// Unwrap implements errors's Unwrap().
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

// GetFunctionName returns the declaration name of a function.
func GetFunctionName(v reflect.Value) string {
	if v.Kind() != reflect.Func {
		return fmt.Sprintf("%T is not a function", v)
	}
	return runtime.FuncForPC(v.Pointer()).Name()
}
