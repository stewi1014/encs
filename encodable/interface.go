package encodable

import (
	"fmt"
	"io"
	"reflect"
	"unsafe"

	"github.com/stewi1014/encs/encio"
)

// Encode calls enc.Encode with the address of the type inside the interface v, checking for type equality.
func Encode(enc Encodable, v interface{}, w io.Writer) error {
	iv := reflect.TypeOf(v)
	if iv != enc.Type() {
		return encio.NewError(encio.ErrBadType, fmt.Sprintf("cannot encode %v with %v Encodable", iv, enc.Type()), 0)
	}
	iptr := ptrInterface(unsafe.Pointer(&v))
	if iv.Kind() == reflect.Ptr {
		return enc.Encode(unsafe.Pointer(&iptr.elem), w)
	}
	return enc.Encode(iptr.elem, w)
}

// Decode creates a new value of the Encodable's type, calls Decode() with it,
// and sets the interface at v to the new value.
func Decode(enc Encodable, v *interface{}, r io.Reader) error {
	if v == nil {
		return encio.NewError(encio.ErrNilPointer, "cannot decode into interface when pointer to it is nil", 0)
	}
	ival := reflect.ValueOf(v).Elem()
	val := reflect.New(enc.Type()).Elem()

	err := enc.Decode(unsafe.Pointer(val.UnsafeAddr()), r)
	ival.Set(val)
	return err
}
