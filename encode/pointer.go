package encode

import (
	"fmt"
	"io"
	"reflect"
	"unsafe"

	"github.com/stewi1014/encs/encio"
)

// NewPointer returns a new pointer Encodable.
func NewPointer(ty reflect.Type, src Source) Encodable {
	if ty.Kind() != reflect.Ptr {
		panic(encio.NewError(encio.ErrBadType, fmt.Sprintf("%v is not a pointer", ty), 0))
	}

	return &Pointer{
		buff: make([]byte, 1),
		ty:   ty,
		elem: src.NewEncodable(ty.Elem(), nil),
	}
}

// Pointer encodes pointers to concrete types.
// It does not follow different semantics to other Encodables,
// and so must be given a non-nil pointer to the pointer it's encoding or decoding.
// If the underlying pointer is nil this is handled as it should be.
type Pointer struct {
	ty   reflect.Type
	elem *Encodable
	buff []byte
}

// Size implements Sized.
func (e *Pointer) Size() int {
	return (*e.elem).Size() + 1
}

// Type implements Encodable.
func (e *Pointer) Type() reflect.Type {
	return e.ty
}

// Encode implements Encodable.
func (e *Pointer) Encode(ptr unsafe.Pointer, w io.Writer) error {
	checkPtr(ptr)
	if *(*unsafe.Pointer)(ptr) == nil {
		// Nil pointer
		e.buff[0] = 1
		return encio.Write(e.buff, w)
	}

	e.buff[0] = 0
	if err := encio.Write(e.buff, w); err != nil {
		return err
	}

	return (*e.elem).Encode(*(*unsafe.Pointer)(ptr), w)
}

// Decode implements Encodable.
func (e *Pointer) Decode(ptr unsafe.Pointer, r io.Reader) error {
	checkPtr(ptr)

	err := encio.Read(e.buff, r)
	if err != nil {
		return err
	}

	if e.buff[0] != 0 {
		// Nil pointer
		v := reflect.NewAt(e.ty, ptr).Elem()
		v.Set(reflect.New(e.ty).Elem())
		return nil
	}

	if *(*unsafe.Pointer)(ptr) == nil {
		reflect.NewAt(e.ty, ptr).Elem().Set(reflect.New(e.ty.Elem()))
	}

	eptr := *(*unsafe.Pointer)(ptr)
	if err := (*e.elem).Decode(eptr, r); err != nil {
		return err
	}

	return nil
}
