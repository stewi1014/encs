package encode

import (
	"io"
	"reflect"
	"unsafe"

	"github.com/stewi1014/encs/encio"
	"github.com/stewi1014/encs/encodable"
	"github.com/stewi1014/encs/types"
)

// NewValue returns a new reflect.Value encodable.
func NewValue(src encodable.Source) *Value {
	return &Value{
		typeEnc: src.NewEncodable(types.ReflectTypeType, nil),
		src:     encodable.NewCachingSource(src),
		buff:    make([]byte, 1),
	}
}

// Value is an encodable for reflect.Value values.
type Value struct {
	typeEnc *encodable.Encodable
	src     *encodable.CachingSource
	buff    []byte
}

const (
	valueValid = 1 << iota
)

// Size implements Encodable.
func (e *Value) Size() int { return -1 << 31 }

// Type implements Encodable.
func (e *Value) Type() reflect.Type { return types.ReflectValueType }

// Encode implements Encodable.
func (e *Value) Encode(ptr unsafe.Pointer, w io.Writer) error {
	checkPtr(ptr)
	v := *(*reflect.Value)(ptr)

	if !v.IsValid() {
		e.buff[0] = 0
		return encio.Write(e.buff, w)
	}
	e.buff[0] = valueValid
	if err := encio.Write(e.buff, w); err != nil {
		return err
	}

	ty := v.Type()

	if !v.CanAddr() {
		// We can safely copy this value, as reflect informs us that this value is unaddressable.
		// reflect really better not be lying to us...
		// If it actually is addressable and reflect just doesn't want us to have the address then we may have a problem.
		// When encoding references I need one of two things. The address to compare with others, or a 100% guarantee that no other pointer *could* exist.
		// A user could always do some weird things (like me :\) with pointers and have a reference that shouldn't be possible in the go type system,
		// but they can write their own Encodable if that's the case.
		n := reflect.New(ty).Elem()
		n.Set(v)
		v = n
	}

	enc := e.src.NewEncodable(ty, nil)

	if err := (*e.typeEnc).Encode(unsafe.Pointer(&ty), w); err != nil {
		return err
	}

	return (*enc).Encode(unsafe.Pointer(v.UnsafeAddr()), w)
}

// Decode implements Encodable.
func (e *Value) Decode(ptr unsafe.Pointer, r io.Reader) error {
	checkPtr(ptr)

	if err := encio.Read(e.buff, r); err != nil {
		return err
	}

	if e.buff[0]&valueValid == 0 {
		*(*reflect.Value)(ptr) = reflect.Value{}
		return nil
	}

	var ty reflect.Type
	if err := (*e.typeEnc).Decode(unsafe.Pointer(&ty), r); err != nil {
		return err
	}

	enc := e.src.NewEncodable(ty, nil)

	v := (*reflect.Value)(ptr)
	if !v.CanAddr() || v.Type() != ty {
		n := reflect.New(ty).Elem()
		if v.IsValid() && ty == v.Type() {
			n.Set(*v)
		}
		*v = n
	}

	if err := (*enc).Decode(unsafe.Pointer(v.UnsafeAddr()), r); err != nil {
		// I no longer trust whatever is in the value
		*(*reflect.Value)(ptr) = reflect.Value{} // zero it
		return err
	}
	return nil
}
