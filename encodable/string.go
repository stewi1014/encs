package encodable

import (
	"fmt"
	"io"
	"reflect"
	"unsafe"

	"github.com/stewi1014/encs/encio"
)

// NewString returns a new string Encodable.
func NewString(ty reflect.Type) *String {
	if ty.Kind() != reflect.String {
		panic(encio.NewError(encio.ErrBadType, fmt.Sprintf("%v is not of string kind", ty.String()), 0))
	}

	return &String{
		ty:  ty,
		len: encio.NewVaruint32(),
	}
}

// String is an Encodable for strings.
type String struct {
	ty  reflect.Type
	len encio.Varuint32
}

// Size implemenets Encodable.
func (e *String) Size() int { return -1 << 31 }

// Type implements Encodable.
func (e *String) Type() reflect.Type { return e.ty }

// Encode implemenets Encodable.
func (e *String) Encode(ptr unsafe.Pointer, w io.Writer) error {
	checkPtr(ptr)
	str := (*string)(ptr)
	l := uint32(len(*str))

	if _, err := e.len.Encode(w, l); err != nil || l == 0 {
		return err
	}

	return encio.Write([]byte(*str), w)
}

// Decode implemenets Encodable.
func (e *String) Decode(ptr unsafe.Pointer, r io.Reader) error {
	checkPtr(ptr)

	l, err := e.len.Decode(r)
	if err != nil {
		return err
	}

	if uintptr(l) > encio.TooBig {
		return encio.NewIOError(
			encio.ErrMalformed,
			r,
			fmt.Sprintf("string with length %v is too big", l),
			0,
		)
	}

	// Create the buffer to hold the string.
	buff := make([]byte, l)
	if err := encio.Read(buff, r); err != nil {
		return err
	}

	*(*string)(ptr) = string(buff)
	return nil
}
