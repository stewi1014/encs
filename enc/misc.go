package enc

import (
	"encoding"
	"fmt"
	"io"
	"reflect"
	"unsafe"
)

// NewString returns a new string Encodable
func NewString() *String {
	return &String{
		buff: make([]byte, 4),
	}
}

// String is an Encodable for strings
type String struct {
	buff []byte
}

func (e *String) String() string {
	return "String"
}

// Size implemenets Encodable
func (e *String) Size() int {
	return -1 << 31
}

// Type implements Encodable
func (e *String) Type() reflect.Type {
	return stringType
}

// Encode implemenets Encodable
func (e *String) Encode(ptr unsafe.Pointer, w io.Writer) error {
	if ptr == nil {
		return ErrNilPointer
	}
	strPtr := ptrString(ptr)
	l := uint32(strPtr.len)
	e.buff[0] = uint8(l)
	e.buff[1] = uint8(l >> 8)
	e.buff[2] = uint8(l >> 16)
	e.buff[3] = uint8(l >> 24)
	if err := write(e.buff, w); err != nil {
		return err
	}

	return write(strPtr.byteSlice(), w)
}

// Decode implemenets Encodable
func (e *String) Decode(ptr unsafe.Pointer, r io.Reader) error {
	if ptr == nil {
		return ErrNilPointer
	}
	if err := read(e.buff, r); err != nil {
		return err
	}

	l := uint32(e.buff[0])
	l |= uint32(e.buff[1]) << 8
	l |= uint32(e.buff[2]) << 16
	l |= uint32(e.buff[3]) << 24
	if int(l) > TooBig {
		return fmt.Errorf("%v; received string length too large", ErrMalformed)
	}

	// I would like to re-use the existing string, but doing so sometimes panics upon writing to the old string.
	// This is probably for the best.
	buff := make([]byte, l)
	if err := read(buff, r); err != nil {
		return err
	}

	// slices and string share the same format up until the end of the string type.
	// so we can just do this, buff isn't going to be used again.
	*(*stringPtr)(ptr) = *(*stringPtr)(unsafe.Pointer(&buff))

	return nil
}

// NewBool returns a new bool Encodable
func NewBool() *Bool {
	return &Bool{
		buff: make([]byte, 1),
	}
}

// Bool is an Encodable for bools
type Bool struct {
	buff []byte
}

func (e *Bool) String() string {
	return "Bool"
}

// Size implements Sized
func (e *Bool) Size() int {
	return 1
}

// Type implements Encodable
func (e *Bool) Type() reflect.Type {
	return boolType
}

// Encode implements Encodable
func (e *Bool) Encode(ptr unsafe.Pointer, w io.Writer) error {
	if ptr == nil {
		return ErrNilPointer
	}
	e.buff[0] = *(*byte)(ptr)
	return write(e.buff, w)
}

// Decode implements Encodable
func (e *Bool) Decode(ptr unsafe.Pointer, r io.Reader) error {
	if ptr == nil {
		return ErrNilPointer
	}
	if err := read(e.buff, r); err != nil {
		return err
	}
	*(*byte)(ptr) = e.buff[0]
	return nil
}

// NewBinaryMarshaler returns a new BinaryMarshaler Encodable.
// It can internally handle a reference;
// i.e. time.Time's unmarshal function requires a reference, but both
// time.Time and *time.Time will function here, as long at ptr is *time.Time or **time.Time respectively.
func NewBinaryMarshaler(t reflect.Type) *BinaryMarshaler {
	e := &BinaryMarshaler{
		t: t,
	}

	err := implementsBinaryMarshaler(t)
	if err != nil {
		if implementsBinaryMarshaler(reflect.PtrTo(t)) != nil {
			panic(err)
		}
		//init referenced
		e.createReference = true
		ival := reflect.ValueOf(&e.i).Elem()
		ival.Set(reflect.New(t))

	} else {
		//init direct
		ival := reflect.ValueOf(&e.i).Elem()
		ival.Set(reflect.New(t).Elem())
	}

	return e
}

// BinaryMarshaler is an Encodable for types which implement encoding.BinaryMarshaler and encoding.BinaryUnmarshaler.
type BinaryMarshaler struct {
	t               reflect.Type
	i               binaryMarshaler
	createReference bool
	buff            [4]byte
}

type binaryMarshaler interface {
	encoding.BinaryMarshaler
	encoding.BinaryUnmarshaler
}

func implementsBinaryMarshaler(t reflect.Type) error {
	if !t.Implements(binaryMarshalerIface) {
		return fmt.Errorf("%v does not implement encoding.BinaryMarshaler", t)
	}
	if !t.Implements(binaryUnmarshalerIface) {
		return fmt.Errorf("%v does not implement encoding.BinaryUnarshaler", t)
	}
	return nil
}

func (e *BinaryMarshaler) setIface(ptr unsafe.Pointer) {
	if e.createReference {
		reflect.ValueOf(&e.i).Elem().Set(reflect.NewAt(e.t, ptr))
		return
	}
	reflect.ValueOf(&e.i).Elem().Set(reflect.NewAt(e.t, ptr).Elem())
}

func (e *BinaryMarshaler) String() string {
	return fmt.Sprintf("BinaryMarshaler(%v)", e.t)
}

// Type implements Encodable
func (e *BinaryMarshaler) Type() reflect.Type {
	return e.t
}

// Size implements Encodable
func (e *BinaryMarshaler) Size() int {
	return -1 << 31
}

// Encode implements Encodable
func (e *BinaryMarshaler) Encode(ptr unsafe.Pointer, w io.Writer) error {
	if ptr == nil {
		return ErrNilPointer
	}

	e.setIface(ptr)

	buff, err := e.i.MarshalBinary()
	if err != nil {
		return err
	}

	l := uint32(len(buff))
	e.buff[0] = uint8(l)
	e.buff[1] = uint8(l >> 8)
	e.buff[2] = uint8(l >> 16)
	e.buff[3] = uint8(l >> 24)
	if err := write(e.buff[:], w); err != nil {
		return err
	}

	return write(buff, w)
}

// Decode implements Encodable
func (e *BinaryMarshaler) Decode(ptr unsafe.Pointer, r io.Reader) error {
	if ptr == nil {
		return ErrNilPointer
	}

	if err := read(e.buff[:], r); err != nil {
		return err
	}

	l := uint32(e.buff[0])
	l |= uint32(e.buff[1]) << 8
	l |= uint32(e.buff[2]) << 16
	l |= uint32(e.buff[3]) << 24
	if int(l) > TooBig {
		return fmt.Errorf("%v: reported BinaryMarshaler buffer size is too large", ErrMalformed)
	}

	buff := make([]byte, l)
	if err := read(buff, r); err != nil {
		return err
	}

	e.setIface(ptr)

	return e.i.UnmarshalBinary(buff)
}
