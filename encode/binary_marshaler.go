package encode

import (
	"encoding"
	"fmt"
	"io"
	"reflect"
	"unsafe"

	"github.com/stewi1014/encs/encio"
	"github.com/stewi1014/encs/types"
)

// NewBinaryMarshaler returns a new BinaryMarshaler Encodable.
// It can internally handle a reference;
// i.e. time.Time's unmarshal function requires a reference, but both
// A type of time.Time and *time.Time will function here, as long as ptr in Encode() and Decode() is *time.Time or **time.Time respectively.
func NewBinaryMarshaler(t reflect.Type) *BinaryMarshaler {
	e := &BinaryMarshaler{
		t: t,
	}

	err := types.ImplementsBinaryMarshaler(t)
	if err != nil {
		if types.ImplementsBinaryMarshaler(reflect.PtrTo(t)) != nil {
			panic(err)
		}
		// init referenced
		e.createReference = true
		ival := reflect.ValueOf(&e.i).Elem()
		ival.Set(reflect.New(t))

	} else {
		// init direct
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

func (e *BinaryMarshaler) setIface(ptr unsafe.Pointer) {
	if e.createReference {
		e.i = reflect.NewAt(e.t, ptr).Interface().(binaryMarshaler)
		return
	}
	e.i = reflect.NewAt(e.t, ptr).Elem().Interface().(binaryMarshaler)
}

// Type implements Encodable.
func (e *BinaryMarshaler) Type() reflect.Type { return e.t }

// Size implements Encodable.
func (e *BinaryMarshaler) Size() int { return -1 << 31 }

// Encode implements Encodable.
func (e *BinaryMarshaler) Encode(ptr unsafe.Pointer, w io.Writer) error {
	checkPtr(ptr)

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

// Decode implements Encodable.
func (e *BinaryMarshaler) Decode(ptr unsafe.Pointer, r io.Reader) error {
	checkPtr(ptr)

	if err := encio.Read(e.buff[:], r); err != nil {
		return err
	}

	l := uint32(e.buff[0])
	l |= uint32(e.buff[1]) << 8
	l |= uint32(e.buff[2]) << 16
	l |= uint32(e.buff[3]) << 24
	if uintptr(l) > encio.TooBig {
		return encio.NewIOError(encio.ErrMalformed, r, fmt.Sprintf("buffer with length %v is too big", l), 0)
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
