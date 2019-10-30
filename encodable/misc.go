package encodable

import (
	"encoding"
	"fmt"
	"io"
	"reflect"
	"unsafe"

	"github.com/stewi1014/encs/encio"
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

// String implements Encodable
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
		return encio.Error{
			Err:    encio.ErrNilPointer,
			Caller: "enc.String.Encode",
		}
	}
	strPtr := ptrString(ptr)
	l := uint32(strPtr.len)
	e.buff[0] = uint8(l)
	e.buff[1] = uint8(l >> 8)
	e.buff[2] = uint8(l >> 16)
	e.buff[3] = uint8(l >> 24)
	if err := encio.Write(e.buff, w); err != nil {
		return err
	}

	return encio.Write(strPtr.byteSlice(), w)
}

// Decode implemenets Encodable
func (e *String) Decode(ptr unsafe.Pointer, r io.Reader) error {
	if ptr == nil {
		return encio.Error{
			Err:    encio.ErrNilPointer,
			Caller: "enc.String.Decode",
		}
	}
	if err := encio.Read(e.buff, r); err != nil {
		return err
	}

	l := uint32(e.buff[0])
	l |= uint32(e.buff[1]) << 8
	l |= uint32(e.buff[2]) << 16
	l |= uint32(e.buff[3]) << 24
	if int(l) > encio.TooBig {
		return encio.IOError{
			Err:     encio.ErrMalformed,
			Message: fmt.Sprintf("string with length %v is too big", l),
		}
	}

	// I would like to re-use the existing string, but doing so sometimes panics upon writing to the old string.
	// This is probably for the best.
	// TODO; revisit this after testing, the issue may not have been caused here.
	buff := make([]byte, l)
	if err := encio.Read(buff, r); err != nil {
		return err
	}

	// slices and strings share the same format up until the end of the string type.
	// so we can just do this, buff isn't going to be used again.
	*(*stringPtr)(ptr) = *(*stringPtr)(unsafe.Pointer(&buff))

	return nil
}

// NewBool returns a new bool Encodable
func NewBool() Encodable {
	return &Bool{
		buff: make([]byte, 1),
	}
}

// Bool is an Encodable for bools
type Bool struct {
	buff []byte
}

// String implements Encodable
func (e *Bool) String() string {
	return "Bool"
}

// Size implements Encodable
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
		return encio.Error{
			Err:    encio.ErrNilPointer,
			Caller: "enc.Bool.Encode",
		}
	}
	e.buff[0] = *(*byte)(ptr)
	return encio.Write(e.buff, w)
}

// Decode implements Encodable
func (e *Bool) Decode(ptr unsafe.Pointer, r io.Reader) error {
	if ptr == nil {
		return encio.Error{
			Err:    encio.ErrNilPointer,
			Caller: "enc.Bool.Decode",
		}
	}
	if err := encio.Read(e.buff, r); err != nil {
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
	mbuff           []byte
}

type binaryMarshaler interface {
	encoding.BinaryMarshaler
	encoding.BinaryUnmarshaler
}

func implementsBinaryMarshaler(t reflect.Type) error {
	if !t.Implements(binaryMarshalerIface) {
		return encio.Error{
			Err:     encio.ErrBadType,
			Caller:  "enc.NewBinaryMarshaler",
			Message: fmt.Sprintf("%v does not implement encoding.BinaryMarshaler", t),
		}
	}
	if !t.Implements(binaryUnmarshalerIface) {
		return encio.Error{
			Err:     encio.ErrBadType,
			Caller:  "enc.NewBinaryMarshaler",
			Message: fmt.Sprintf("%v does not implement encoding.BinaryUnmarshaler", t),
		}
	}
	return nil
}

func (e *BinaryMarshaler) setIface(ptr unsafe.Pointer) {
	if e.createReference {
		e.i = reflect.NewAt(e.t, ptr).Interface().(binaryMarshaler)
		return
	}
	e.i = reflect.NewAt(e.t, ptr).Elem().Interface().(binaryMarshaler)
}

// String implements Encodable
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
		return encio.Error{
			Err:    encio.ErrNilPointer,
			Caller: "enc.BinaryMarshaler.Encode",
		}
	}

	e.setIface(ptr)

	var err error
	e.mbuff, err = e.i.MarshalBinary()
	if err != nil {
		return err
	}

	l := uint32(len(e.mbuff))
	e.buff[0] = uint8(l)
	e.buff[1] = uint8(l >> 8)
	e.buff[2] = uint8(l >> 16)
	e.buff[3] = uint8(l >> 24)
	if err := encio.Write(e.buff[:], w); err != nil {
		return err
	}

	return encio.Write(e.mbuff, w)
}

// Decode implements Encodable
func (e *BinaryMarshaler) Decode(ptr unsafe.Pointer, r io.Reader) error {
	if ptr == nil {
		return encio.Error{
			Err:    encio.ErrNilPointer,
			Caller: "enc.BinaryMarshaler.Decode",
		}
	}

	if err := encio.Read(e.buff[:], r); err != nil {
		return err
	}

	l := uint32(e.buff[0])
	l |= uint32(e.buff[1]) << 8
	l |= uint32(e.buff[2]) << 16
	l |= uint32(e.buff[3]) << 24
	if int(l) > encio.TooBig {
		return encio.IOError{
			Err:     encio.ErrMalformed,
			Message: fmt.Sprintf("buffer with length %v is too big", l),
		}
	}

	if cap(e.mbuff) < int(l) {
		e.mbuff = make([]byte, l)
	}
	e.mbuff = e.mbuff[:l]
	if err := encio.Read(e.mbuff, r); err != nil {
		return err
	}

	e.setIface(ptr)

	return e.i.UnmarshalBinary(e.mbuff)
}
