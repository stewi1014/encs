package encodable

import (
	"fmt"
	"io"
	"reflect"
	"unsafe"

	"github.com/stewi1014/encs/encio"
)

// NewUint8 returns a new uint8 Encodable.
func NewUint8(ty reflect.Type) *Uint8 {
	if ty.Kind() != reflect.Uint8 {
		panic(encio.NewError(encio.ErrBadType, fmt.Sprintf("%v is not of uint8 kind", ty.String()), 0))
	}
	return &Uint8{
		ty: ty,
	}
}

// Uint8 is an Encodable for uint8s.
type Uint8 struct {
	ty   reflect.Type
	buff [1]byte
}

// Size implements Encodable.
func (e *Uint8) Size() int { return 1 }

// Type implements Encodable.
func (e *Uint8) Type() reflect.Type { return e.ty }

// Encode implements Encodable.
func (e *Uint8) Encode(ptr unsafe.Pointer, w io.Writer) error {
	checkPtr(ptr)
	e.buff[0] = *(*uint8)(ptr)
	return encio.Write(e.buff[:], w)
}

// Decode implements Encodable.
func (e *Uint8) Decode(ptr unsafe.Pointer, r io.Reader) error {
	checkPtr(ptr)
	if err := encio.Read(e.buff[:], r); err != nil {
		return err
	}

	*(*uint8)(ptr) = e.buff[0]
	return nil
}

// NewUint16 returns a new uint16 Encodable.
func NewUint16(ty reflect.Type) *Uint16 {
	if ty.Kind() != reflect.Uint16 {
		panic(encio.NewError(encio.ErrBadType, fmt.Sprintf("%v is not of uint16 kind", ty.String()), 0))
	}
	return &Uint16{
		ty: ty,
	}
}

// Uint16 is an Encodable for uint16s.
type Uint16 struct {
	ty   reflect.Type
	buff [2]byte
}

// Size implements Encodable.
func (e *Uint16) Size() int { return 2 }

// Type implements Encodable.
func (e Uint16) Type() reflect.Type { return e.ty }

// Encode implements Encodable.
func (e *Uint16) Encode(ptr unsafe.Pointer, w io.Writer) error {
	checkPtr(ptr)
	i := *(*uint16)(ptr)
	e.buff[0] = uint8(i)
	e.buff[1] = uint8(i >> 8)
	return encio.Write(e.buff[:], w)
}

// Decode implements Encodable.
func (e *Uint16) Decode(ptr unsafe.Pointer, r io.Reader) error {
	checkPtr(ptr)
	if err := encio.Read(e.buff[:], r); err != nil {
		return err
	}

	i := (*uint16)(ptr)
	*i = uint16(e.buff[0])
	*i |= uint16(e.buff[1]) << 8
	return nil
}

// NewUint32 returns a new uint32 Encodable.
func NewUint32(ty reflect.Type) *Uint32 {
	if ty.Kind() != reflect.Uint32 {
		panic(encio.NewError(encio.ErrBadType, fmt.Sprintf("%v is not of uint32 kind", ty.String()), 0))
	}
	return &Uint32{
		ty: ty,
	}
}

// Uint32 is an Encodable for uint32s.
type Uint32 struct {
	ty   reflect.Type
	buff [4]byte
}

// Size implements Encodable.
func (e *Uint32) Size() int { return 4 }

// Type implements Encodable.
func (e *Uint32) Type() reflect.Type { return e.ty }

// Encode implements Encodable.
func (e *Uint32) Encode(ptr unsafe.Pointer, w io.Writer) error {
	checkPtr(ptr)
	i := *(*uint32)(ptr)
	e.buff[0] = uint8(i)
	e.buff[1] = uint8(i >> 8)
	e.buff[2] = uint8(i >> 16)
	e.buff[3] = uint8(i >> 24)
	return encio.Write(e.buff[:], w)
}

// Decode implements Encodable.
func (e *Uint32) Decode(ptr unsafe.Pointer, r io.Reader) error {
	checkPtr(ptr)
	err := encio.Read(e.buff[:], r)
	if err != nil {
		return err
	}

	i := (*uint32)(ptr)
	*i = uint32(e.buff[0])
	*i |= uint32(e.buff[1]) << 8
	*i |= uint32(e.buff[2]) << 16
	*i |= uint32(e.buff[3]) << 24
	return nil
}

// NewUint64 returns a new uint64 Encodable.
func NewUint64(ty reflect.Type) *Uint64 {
	if ty.Kind() != reflect.Uint64 {
		panic(encio.NewError(encio.ErrBadType, fmt.Sprintf("%v is not of uint64 kind", ty.String()), 0))
	}
	return &Uint64{
		ty: ty,
	}
}

// Uint64 is an Encodable for uint64s.
type Uint64 struct {
	ty   reflect.Type
	buff [8]byte
}

// Size implements Encodable.
func (e *Uint64) Size() int { return 8 }

// Type implements Encodable.
func (e *Uint64) Type() reflect.Type { return e.ty }

// Encode implements Encodable.
func (e *Uint64) Encode(ptr unsafe.Pointer, w io.Writer) error {
	checkPtr(ptr)
	i := *(*uint64)(ptr)
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
func (e *Uint64) Decode(ptr unsafe.Pointer, r io.Reader) error {
	checkPtr(ptr)
	err := encio.Read(e.buff[:], r)
	if err != nil {
		return err
	}

	i := (*uint64)(ptr)
	*i = uint64(e.buff[0])
	*i |= uint64(e.buff[1]) << 8
	*i |= uint64(e.buff[2]) << 16
	*i |= uint64(e.buff[3]) << 24
	*i |= uint64(e.buff[4]) << 32
	*i |= uint64(e.buff[5]) << 40
	*i |= uint64(e.buff[6]) << 48
	*i |= uint64(e.buff[7]) << 56
	return nil
}

// NewUint returns a new uint Encodable.
func NewUint(ty reflect.Type) *Uint {
	if ty.Kind() != reflect.Uint {
		panic(encio.NewError(encio.ErrBadType, fmt.Sprintf("%v is not of uint kind", ty.String()), 0))
	}
	return &Uint{
		ty: ty,
	}
}

// Uint is an Encodable for uints.
type Uint struct {
	ty   reflect.Type
	buff [9]byte
}

const (
	maxSingleUint = 255 - 8
)

// Size implements Encodable.
func (e *Uint) Size() int { return 9 }

// Type implements Encodable.
func (e *Uint) Type() reflect.Type { return e.ty }

// Encode implements Encodable.
func (e *Uint) Encode(ptr unsafe.Pointer, w io.Writer) error {
	checkPtr(ptr)
	i := *(*uint)(ptr)
	size := uint8(1)
	if i <= maxSingleUint {
		e.buff[0] = uint8(i)
	} else {
		for i > 0 {
			e.buff[size] = uint8(i)
			i >>= 8
			size++
		}
		e.buff[0] = maxSingleUint + size - 1
	}

	return encio.Write(e.buff[:size], w)
}

// Decode implements Encodable.
func (e *Uint) Decode(ptr unsafe.Pointer, r io.Reader) error {
	checkPtr(ptr)
	if err := encio.Read(e.buff[:1], r); err != nil {
		return err
	}

	i := (*uint)(ptr)

	if e.buff[0] > maxSingleUint {
		size := e.buff[0] - maxSingleUint
		if err := encio.Read(e.buff[:size], r); err != nil {
			return err
		}

		*i = 0
		for j := byte(0); j < size; j++ {
			*i |= uint(e.buff[j]) << (j * 8)
		}
	} else {
		*i = uint(e.buff[0])
	}

	return nil
}

// NewUintptr returns a new uintptr Encodable.
func NewUintptr(ty reflect.Type) *Uintptr {
	if ty.Kind() != reflect.Uintptr {
		panic(encio.NewError(encio.ErrBadType, fmt.Sprintf("%v is not of uintptr kind", ty.String()), 0))
	}
	return &Uintptr{
		ty: ty,
	}
}

// Uintptr is an Encodable for uintptrs.
type Uintptr struct {
	ty   reflect.Type
	buff [9]byte
}

// Size implements Encodable.
func (e *Uintptr) Size() int { return 9 }

// Type implements Encodable.
func (e *Uintptr) Type() reflect.Type { return e.ty }

// Encode implements Encodable.
func (e *Uintptr) Encode(ptr unsafe.Pointer, w io.Writer) error {
	checkPtr(ptr)
	i := *(*uintptr)(ptr)
	l := uint8(1)
	if i <= maxSingleUint {
		e.buff[0] = uint8(i)
	} else {
		for i > 0 {
			e.buff[l] = uint8(i)
			i >>= 8
			l++
		}
		e.buff[0] = maxSingleUint + l - 1
	}

	return encio.Write(e.buff[:l], w)
}

// Decode implements Encodable.
func (e *Uintptr) Decode(ptr unsafe.Pointer, r io.Reader) error {
	checkPtr(ptr)
	if err := encio.Read(e.buff[:1], r); err != nil {
		return err
	}

	i := (*uintptr)(ptr)

	if e.buff[0] > maxSingleUint {
		size := e.buff[0] - maxSingleUint
		if err := encio.Read(e.buff[:size], r); err != nil {
			return err
		}

		*i = 0
		for j := byte(0); j < size; j++ {
			*i |= uintptr(e.buff[j]) << (j * 8)
		}
	} else {
		*i = uintptr(e.buff[0])
	}

	return nil
}
