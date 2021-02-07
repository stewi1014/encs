package encode

import (
	"fmt"
	"io"
	"reflect"
	"unsafe"

	"github.com/stewi1014/encs/encio"
)

// NewBool returns a new bool Encodable.
func NewBool(ty reflect.Type) *Bool {
	if ty.Kind() != reflect.Bool {
		panic(encio.NewError(encio.ErrBadType, fmt.Sprintf("%v is not of bool kind", ty.String()), 0))
	}
	return &Bool{
		ty:   ty,
		buff: make([]byte, 1),
	}
}

// Bool is an Encodable for bools.
type Bool struct {
	ty   reflect.Type
	buff []byte
}

// Size implements Encodable.
func (e *Bool) Size() int { return 1 }

// Type implements Encodable.
func (e *Bool) Type() reflect.Type { return e.ty }

// Encode implements Encodable.
func (e *Bool) Encode(ptr unsafe.Pointer, w io.Writer) error {
	checkPtr(ptr)
	e.buff[0] = *(*byte)(ptr)
	return encio.Write(e.buff, w)
}

// Decode implements Encodable.
func (e *Bool) Decode(ptr unsafe.Pointer, r io.Reader) error {
	checkPtr(ptr)
	if err := encio.Read(e.buff, r); err != nil {
		return err
	}
	*(*byte)(ptr) = e.buff[0]
	return nil
}
