package enc

import (
	"io"
	"reflect"
	"unsafe"
)

// EncodeInterface encodes the value referenced by v with e.
// If v's type is not e's type, it returns ErrBadType
func EncodeInterface(i interface{}, e Encodable, w io.Writer) error {
	if reflect.TypeOf(i) != e.Type() {
		return ErrBadType
	}

	iptr := ptrInterface(unsafe.Pointer(&i))
	return e.Encode(iptr.ptr(), w)
}

// DecodeInterface decodes the value read from r into the interface.
// If the interface is nil or contains a different type, a new type is allocated.
func DecodeInterface(i *interface{}, e Encodable, r io.Reader) error {
	if i == nil {
		return ErrNilPointer
	}

	ival := reflect.ValueOf(i).Elem()
	if ival.IsNil() || ival.Elem().Type() != e.Type() {
		ival.Set(reflect.New(e.Type()).Elem())
	}

	iptr := ptrInterface(unsafe.Pointer(i))
	return e.Decode(iptr.ptr(), r)
}
