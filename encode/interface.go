package encode

import (
	"fmt"
	"io"
	"reflect"
	"unsafe"

	"github.com/stewi1014/encs/encio"
)

// NewInterface returns a new interface Encodable.
func NewInterface(ty reflect.Type, src Source) *Interface {
	if ty.Kind() != reflect.Interface {
		panic(encio.NewError(encio.ErrBadType, fmt.Sprintf("%v is not an interface", ty), 0))
	}

	e := &Interface{
		ty:      ty,
		source:  src,
		typeEnc: src.NewEncodable(reflectTypeType, nil),
		buff:    make([]byte, 1),
	}

	return e
}

// Interface is an Encodable for interfaces.
type Interface struct {
	ty      reflect.Type
	source  Source
	typeEnc *Encodable
	buff    []byte
}

// Size implements Encodable.
func (e *Interface) Size() int {
	return -1 << 31
}

// Type implements Encodable.
func (e *Interface) Type() reflect.Type {
	return e.ty
}

// Encode implements Encodable.
func (e *Interface) Encode(ptr unsafe.Pointer, w io.Writer) error {
	checkPtr(ptr)

	i := reflect.NewAt(e.ty, ptr).Elem()
	if i.IsNil() {
		e.buff[0] = 0
		return encio.Write(e.buff, w)
	}

	e.buff[0] = 1
	err := encio.Write(e.buff, w)
	if err != nil {
		return err
	}

	elemType := i.Elem().Type()
	elem := reflect.New(elemType).Elem()
	elem.Set(i.Elem())

	err = (*e.typeEnc).Encode(unsafe.Pointer(&elemType), w)
	if err != nil {
		return err
	}

	elemEnc := e.source.NewEncodable(elemType, nil)
	return (*elemEnc).Encode(unsafe.Pointer(elem.UnsafeAddr()), w)
}

// Decode implements Encodable.
func (e *Interface) Decode(ptr unsafe.Pointer, r io.Reader) error {
	checkPtr(ptr)

	if err := encio.Read(e.buff, r); err != nil {
		return err
	}

	i := reflect.NewAt(e.ty, ptr).Elem()

	if e.buff[0] == 0 {
		// Nil interface
		i.Set(reflect.New(e.ty).Elem())
		return nil
	}

	var elemt reflect.Type
	if !i.IsNil() {
		elemt = i.Elem().Type()
	}

	var rty reflect.Type
	err := (*e.typeEnc).Decode(unsafe.Pointer(&rty), r)
	if err != nil {
		return err
	}

	elem := reflect.New(rty).Elem()
	if rty == elemt {
		// The interface already holds a value of the same type as we're receiving.
		// Unfortunately we can't simply go in and modify it directly due to complex interface semantics.
		// What we can do however, is
		elem.Set(i.Elem()) // copy the existing value.

		// Now why would we do that?
		// While we're stuck with the previous allocation, **subsequent encodables do not have to be stuck allocating everything after this**.
		// For instance, if the interface was holding a struct with a slice member then when the slice encodable gets around to do its thing,
		// it will find the backing array pointer and avoid allocating a new one, assuming it's non-nil and cap is large enough.
	}

	enc := e.source.NewEncodable(rty, nil)
	if err := (*enc).Decode(unsafe.Pointer(elem.UnsafeAddr()), r); err != nil {
		return err
	}

	if !rty.Implements(e.ty) {
		// If loose typing is enabled, then there's a possibility the decoded type doesn't implement the interface.
		return encio.NewError(
			encio.ErrBadType,
			fmt.Sprintf("%v was sent to us inside the %v interface, but %v does not implement %v!", rty, i.Type(), rty, i.Type()),
			0,
		)
	}

	i.Set(elem)

	return nil
}
