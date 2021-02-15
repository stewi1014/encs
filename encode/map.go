package encode

import (
	"fmt"
	"io"
	"reflect"
	"unsafe"

	"github.com/stewi1014/encs/encio"
	"github.com/stewi1014/encs/encodable"
)

// NewMap returns a new map Encodable.
func NewMap(ty reflect.Type, src encodable.Source) *Map {
	if ty.Kind() != reflect.Map {
		panic(encio.NewError(encio.ErrBadType, fmt.Sprintf("%v is not a map", ty), 0))
	}

	return &Map{
		key: src.NewEncodable(ty.Key(), nil),
		val: src.NewEncodable(ty.Elem(), nil),
		len: encio.NewInt32(),
		t:   ty,
	}
}

// Map is an Encodable for maps.
type Map struct {
	key, val *encodable.Encodable
	len      encio.Int32
	t        reflect.Type
}

// Size implements Encodable.
func (e *Map) Size() int { return -1 }

// Type implements Encodable.
func (e *Map) Type() reflect.Type {
	return e.t
}

// Encode implements Encodable.
func (e *Map) Encode(ptr unsafe.Pointer, w io.Writer) error {
	checkPtr(ptr)
	v := reflect.NewAt(e.t, ptr).Elem()

	if v.IsNil() {
		return e.len.Encode(w, -1)
	}

	if err := e.len.Encode(w, int32(v.Len())); err != nil {
		return err
	}

	key := reflect.New(e.t.Key()).Elem()
	val := reflect.New(e.t.Elem()).Elem()

	iter := v.MapRange()
	for iter.Next() {
		key.Set(iter.Key())
		val.Set(iter.Value())

		err := (*e.key).Encode(unsafe.Pointer(key.UnsafeAddr()), w)
		if err != nil {
			return err
		}

		err = (*e.val).Encode(unsafe.Pointer(val.UnsafeAddr()), w)
		if err != nil {
			return err
		}
	}

	return nil
}

// Decode implements Encodable.
func (e *Map) Decode(ptr unsafe.Pointer, r io.Reader) error {
	checkPtr(ptr)
	l, err := e.len.Decode(r)
	if err != nil {
		return err
	}

	m := reflect.NewAt(e.t, ptr).Elem()

	if l < 0 {
		m.Set(reflect.New(e.t).Elem())
		return nil
	}

	if uintptr(l)*((*e.key).Type().Size()+(*e.val).Type().Size()) > encio.TooBig {
		return encio.NewIOError(encio.ErrMalformed, r, fmt.Sprintf("map size of %v is too big", l), 0)
	}

	v := reflect.MakeMapWithSize(e.t, int(l))
	m.Set(v)

	for i := int32(0); i < l; i++ {
		nKey := reflect.New((*e.key).Type()).Elem()
		err := (*e.key).Decode(unsafe.Pointer(nKey.UnsafeAddr()), r)
		if err != nil {
			return err
		}

		nVal := reflect.New((*e.val).Type()).Elem()
		err = (*e.val).Decode(unsafe.Pointer(nVal.UnsafeAddr()), r)
		if err != nil {
			return err
		}

		v.SetMapIndex(nKey, nVal)
	}

	return nil
}
