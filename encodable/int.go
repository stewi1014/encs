package encodable

import (
	"fmt"
	"io"
	"math"
	"reflect"
	"unsafe"

	"github.com/stewi1014/encs/encio"
)

// NewInt8 returns a new int8 Encodable.
func NewInt8(ty reflect.Type) *Int8 {
	if ty.Kind() != reflect.Int8 {
		panic(encio.NewError(encio.ErrBadType, fmt.Sprintf("%v is not of int8 kind", ty.String()), 0))
	}
	return &Int8{
		ty: ty,
	}
}

// Int8 is an Encodable for int8s.
type Int8 struct {
	ty   reflect.Type
	buff [1]byte
}

// Size implements Encodable.
func (e *Int8) Size() int { return 1 }

// Type implements Encodable.
func (e *Int8) Type() reflect.Type { return e.ty }

// Encode implements Encodable.
func (e *Int8) Encode(ptr unsafe.Pointer, w io.Writer) error {
	checkPtr(ptr)
	e.buff[0] = *(*uint8)(ptr)
	return encio.Write(e.buff[:], w)
}

// Decode implements Encodable.
func (e *Int8) Decode(ptr unsafe.Pointer, r io.Reader) error {
	checkPtr(ptr)
	if err := encio.Read(e.buff[:], r); err != nil {
		return err
	}
	*(*int8)(ptr) = int8(e.buff[0])
	return nil
}

// NewInt16 returns a new int16 Encodable.
func NewInt16(ty reflect.Type) *Int16 {
	if ty.Kind() != reflect.Int16 {
		panic(encio.NewError(encio.ErrBadType, fmt.Sprintf("%v is not of int16 kind", ty.String()), 0))
	}
	return &Int16{
		ty: ty,
	}
}

// Int16 is an Encodable for int16s.
type Int16 struct {
	ty   reflect.Type
	buff [2]byte
}

// Size implements Encodable.
func (e *Int16) Size() int { return 2 }

// Type implements Encodable.
func (e *Int16) Type() reflect.Type { return e.ty }

// Encode implements Encodable.
func (e *Int16) Encode(ptr unsafe.Pointer, w io.Writer) error {
	checkPtr(ptr)
	i := *(*int16)(ptr)
	e.buff[0] = uint8(i)
	e.buff[1] = uint8(i >> 8)

	return encio.Write(e.buff[:], w)
}

// Decode implements Encodable.
func (e *Int16) Decode(ptr unsafe.Pointer, r io.Reader) error {
	checkPtr(ptr)
	if err := encio.Read(e.buff[:], r); err != nil {
		return err
	}
	i := (*int16)(ptr)
	*i = int16(e.buff[0])
	*i |= int16(e.buff[1]) << 8
	return nil
}

// NewInt32 returns a new int32 Encodable.
func NewInt32(ty reflect.Type) *Int32 {
	if ty.Kind() != reflect.Int32 {
		panic(encio.NewError(encio.ErrBadType, fmt.Sprintf("%v is not of int32 kind", ty.String()), 0))
	}
	return &Int32{
		ty: ty,
	}
}

// Int32 is an Encodable for int32s.
type Int32 struct {
	ty   reflect.Type
	buff [4]byte
}

// Size implements Encodable.
func (e *Int32) Size() int { return 4 }

// Type implements Encodable.
func (e *Int32) Type() reflect.Type { return e.ty }

// Encode implements Encodable.
func (e *Int32) Encode(ptr unsafe.Pointer, w io.Writer) error {
	checkPtr(ptr)
	i := *(*int32)(ptr)
	e.buff[0] = uint8(i)
	e.buff[1] = uint8(i >> 8)
	e.buff[2] = uint8(i >> 16)
	e.buff[3] = uint8(i >> 24)

	return encio.Write(e.buff[:], w)
}

// Decode implements Encodable.
func (e *Int32) Decode(ptr unsafe.Pointer, r io.Reader) error {
	checkPtr(ptr)
	if err := encio.Read(e.buff[:], r); err != nil {
		return err
	}

	i := (*int32)(ptr)
	*i = int32(e.buff[0])
	*i |= int32(e.buff[1]) << 8
	*i |= int32(e.buff[2]) << 16
	*i |= int32(e.buff[3]) << 24
	return nil
}

// NewInt64 returns a new int64 Encodable.
func NewInt64(ty reflect.Type) *Int64 {
	if ty.Kind() != reflect.Int64 {
		panic(encio.NewError(encio.ErrBadType, fmt.Sprintf("%v is not of int64 kind", ty.String()), 0))
	}
	return &Int64{
		ty: ty,
	}
}

// Int64 is an Encodable for int64s.
type Int64 struct {
	ty   reflect.Type
	buff [8]byte
}

// Size implements Encodable.
func (e *Int64) Size() int { return 8 }

// Type implements Encodable.
func (e *Int64) Type() reflect.Type { return e.ty }

// Encode implements Encodable.
func (e *Int64) Encode(ptr unsafe.Pointer, w io.Writer) error {
	checkPtr(ptr)
	i := *(*int64)(ptr)
	e.buff[0] = uint8(i)
	e.buff[1] = uint8(i >> 8)
	e.buff[2] = uint8(i >> 16)
	e.buff[3] = uint8(i >> 24)
	e.buff[4] = uint8(i >> 32)
	e.buff[5] = uint8(i >> 40)
	e.buff[6] = uint8(i >> 48)
	e.buff[7] = uint8(i >> 56)

	return encio.Write(e.buff[:], w)
}

// Decode implements Encodable.
func (e *Int64) Decode(ptr unsafe.Pointer, r io.Reader) error {
	checkPtr(ptr)
	if err := encio.Read(e.buff[:], r); err != nil {
		return err
	}

	i := (*int64)(ptr)
	*i = int64(e.buff[0])
	*i |= int64(e.buff[1]) << 8
	*i |= int64(e.buff[2]) << 16
	*i |= int64(e.buff[3]) << 24
	*i |= int64(e.buff[4]) << 32
	*i |= int64(e.buff[5]) << 40
	*i |= int64(e.buff[6]) << 48
	*i |= int64(e.buff[7]) << 56
	return nil
}

// NewInt returns a new int Encodable.
func NewInt(ty reflect.Type) *Int {
	if ty.Kind() != reflect.Int {
		panic(encio.NewError(encio.ErrBadType, fmt.Sprintf("%v is not of int kind", ty.String()), 0))
	}
	return &Int{
		ty: ty,
	}
}

// Int is an Encodable for ints.
type Int struct {
	ty   reflect.Type
	buff [9]byte
}

const minSingleInt = int8(-1<<7 + 9)

// Size implements Encodable.
func (e *Int) Size() int { return 9 }

// Type implements Encodable.
func (e *Int) Type() reflect.Type { return e.ty }

// Encode implements Encodable.
func (e *Int) Encode(ptr unsafe.Pointer, w io.Writer) error {
	checkPtr(ptr)
	i := *(*int)(ptr)
	size := 1
	if i <= math.MaxInt8 && i >= int(minSingleInt) {
		e.buff[0] = uint8(i)
	} else {
		end := int(0)
		if i < 0 {
			end = -1
		}

		for i != end {
			e.buff[size] = uint8(i)
			i >>= 8
			size++
		}

		e.buff[0] = uint8((-1 << 7) + size - 1)
	}

	return encio.Write(e.buff[:size], w)
}

// Decode implements Encodable.
func (e *Int) Decode(ptr unsafe.Pointer, r io.Reader) error {
	checkPtr(ptr)

	if err := encio.Read(e.buff[:1], r); err != nil {
		return err
	}

	b := int8(e.buff[0])
	if b >= minSingleInt {
		*(*int)(ptr) = int(b)
		return nil
	}

	size := int(b - (-1 << 7))
	if err := encio.Read(e.buff[:size], r); err != nil {
		return err
	}

	i := (*int)(ptr)
	*i = 0
	for j := 0; j < size; j++ {
		*i |= int(e.buff[j]) << (j * 8)
	}

	return nil
}
