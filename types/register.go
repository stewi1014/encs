package types

import (
	"errors"
	"fmt"
	"reflect"
	"time"
)

func init() {
	err := Register(
		nil,
		reflect.TypeOf(int(0)),
		reflect.TypeOf(int8(0)),
		reflect.TypeOf(int16(0)),
		reflect.TypeOf(int32(0)),
		reflect.TypeOf(int64(0)),
		reflect.TypeOf(uint(0)),
		reflect.TypeOf(uint8(0)),
		reflect.TypeOf(uint16(0)),
		reflect.TypeOf(uint32(0)),
		reflect.TypeOf(uint64(0)),
		reflect.TypeOf(uintptr(0)),
		reflect.TypeOf(float32(0)),
		reflect.TypeOf(float64(0)),
		reflect.TypeOf(complex64(0)),
		reflect.TypeOf(complex128(0)),
		reflect.TypeOf(string("")),
		reflect.TypeOf(bool(false)),
		reflect.TypeOf(time.Time{}),
		reflect.TypeOf(time.Duration(0)),
		ReflectTypeType,
		ReflectValueType,
	)

	if err != nil {
		panic(err)
	}
}

var Registered []reflect.Type

// ErrAlreadyRegistered is returned by Register if one or more of the given types has already been registered.
// It is wrapped.
var ErrAlreadyRegistered = errors.New("already registered")

// Register allows an unknown type to be decoded with Type.
func Register(types ...reflect.Type) error {
	var errMsg string
	for _, ty := range types {
		if !register(ty) {
			if len(errMsg) > 0 {
				errMsg += ", "
			}
			errMsg += fmt.Sprintf("%v", ty)
			continue
		}
	}

	if errMsg != "" {
		return fmt.Errorf("%w: %v", ErrAlreadyRegistered, errMsg)
	}
	return nil
}

// register registers the type, returning false if it was previously registered.
func register(ty reflect.Type) bool {
	for _, r := range Registered {
		if r == ty {
			return false
		}
	}
	Registered = append(Registered, ty)
	return true
}
