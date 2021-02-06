package encode

import (
	"fmt"
	"io"
	"reflect"
	"unsafe"

	"github.com/stewi1014/encs/encio"
)

// NewSlice returns a new slice Encodable.
func NewSlice(ty reflect.Type, src Source) *Slice {
	if ty.Kind() != reflect.Slice {
		panic(encio.NewError(encio.ErrBadType, fmt.Sprintf("%v is not a slice", ty), 0))
	}

	return &Slice{
		t:    ty,
		elem: src.NewEncodable(ty.Elem(), nil),
		len:  encio.NewInt32(),
	}
}

// Slice is an Encodable for slices.
type Slice struct {
	t    reflect.Type
	elem *Encodable
	len  encio.Int32
}

// Size implemenets Encodable.
func (e *Slice) Size() int {
	return -1 << 31
}

// Type implements Encodable.
func (e *Slice) Type() reflect.Type {
	return e.t
}

// Encode implements Encodable.Encode.
// Encoded 0-len and nil slices both have the effect of setting the decoded slice's
// len and cap to 0. nil-ness of the slice being decoded into is retained.
func (e *Slice) Encode(ptr unsafe.Pointer, w io.Writer) error {
	checkPtr(ptr)

	slice := reflect.NewAt(e.t, ptr).Elem()
	if slice.IsNil() {
		return e.len.Encode(w, -1)
	}

	l := slice.Len()
	if err := e.len.Encode(w, int32(l)); err != nil {
		return err
	}

	for i := 0; i < l; i++ {
		err := (*e.elem).Encode(unsafe.Pointer(slice.Index(i).UnsafeAddr()), w)
		if err != nil {
			return err
		}
	}
	return nil
}

// Decode implemenets Encodable.
// Encoded 0-len and nil slices both have the effect of setting the decoded slice's
// len and cap to 0. nil-ness of the slice being decoded into is retained.
func (e *Slice) Decode(ptr unsafe.Pointer, r io.Reader) error {
	checkPtr(ptr)
	slice := reflect.NewAt(e.t, ptr).Elem()

	l, err := e.len.Decode(r)
	if err != nil {
		return err
	}
	if l < 0 {
		// Nil slice
		slice.Set(reflect.New(e.t).Elem())
		return nil
	}

	if uintptr(l)*((*e.elem).Type().Size()) > uintptr(encio.TooBig) {
		return encio.NewIOError(
			encio.ErrMalformed,
			r,
			fmt.Sprintf("slice of length %v (%v bytes) is too big", l, int(l)*int((*e.elem).Type().Size())),
			0,
		)
	}

	length := int(l)
	if slice.Cap() < length || slice.Cap() == 0 {
		// Not enough space, allocate
		slice.Set(reflect.MakeSlice(e.t, length, length))
	} else {
		slice.SetLen(length)
	}

	for i := 0; i < length; i++ {
		eptr := unsafe.Pointer(slice.Index(i).UnsafeAddr())
		err := (*e.elem).Decode(eptr, r)
		if err != nil {
			return err
		}
	}

	return nil
}
