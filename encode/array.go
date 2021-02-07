package encode

import (
	"fmt"
	"io"
	"reflect"
	"unsafe"

	"github.com/stewi1014/encs/encio"
	"github.com/stewi1014/encs/encodable"
)

// NewArray returns a new array Encodable.
func NewArray(ty reflect.Type, src encodable.Source) *Array {
	if ty.Kind() != reflect.Array {
		panic(encio.NewError(encio.ErrBadType, fmt.Sprintf("%v is not an Array", ty), 0))
	}

	return &Array{
		len:  uintptr(ty.Len()),
		size: ty.Elem().Size(),
		elem: src.NewEncodable(ty.Elem(), nil),
	}
}

// Array is an Encodable for arrays.
type Array struct {
	elem *encodable.Encodable
	len  uintptr
	size uintptr
}

// Size implements Encodable.
func (e *Array) Size() int {
	s := (*e.elem).Size()
	if s < 0 {
		return -1 << 31
	}
	return s * int(e.len)
}

// Type implements Encodable.
func (e *Array) Type() reflect.Type {
	return reflect.ArrayOf(int(e.len), (*e.elem).Type())
}

// Encode implements Encodable.
func (e *Array) Encode(ptr unsafe.Pointer, w io.Writer) error {
	checkPtr(ptr)
	for i := uintptr(0); i < e.len; i++ {
		eptr := unsafe.Pointer(uintptr(ptr) + (i * e.size))
		err := (*e.elem).Encode(eptr, w)
		if err != nil {
			return err
		}
	}
	return nil
}

// Decode implments Encodable.
func (e *Array) Decode(ptr unsafe.Pointer, r io.Reader) error {
	checkPtr(ptr)
	for i := uintptr(0); i < e.len; i++ {
		eptr := unsafe.Pointer(uintptr(ptr) + (i * e.size))
		err := (*e.elem).Decode(eptr, r)
		if err != nil {
			return err
		}
	}
	return nil
}
