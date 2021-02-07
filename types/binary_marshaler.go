package types

import (
	"fmt"
	"reflect"

	"github.com/stewi1014/encs/encio"
)

// ImplementsBinaryMarshaler returns a helpful error if the given type does not implement binary marshaler.
func ImplementsBinaryMarshaler(t reflect.Type) error {
	if !t.Implements(BinaryMarshalerType) {
		return encio.NewError(encio.ErrBadType, fmt.Sprintf("%v does not implement encoding.BinaryMarshaler", t), 1)
	}
	if !t.Implements(BinaryUnmarshalerType) {
		return encio.NewError(encio.ErrBadType, fmt.Sprintf("%v does not implement encoding.BinaryUnmarshaler", t), 1)
	}
	return nil
}
