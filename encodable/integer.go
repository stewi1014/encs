package encodable

import (
	"io"
	"math"
	"reflect"
	"unsafe"

	"github.com/stewi1014/encs/encio"
)

// Encoders for basic types.
// These encoders have undefined behaviour if given the wrong type.

// NewUint8 returns a new uint8 Encodable.
func NewUint8() *Uint8 {
	return &Uint8{}
}

// Uint8 is an Encodable for uint8s.
type Uint8 struct {
	buff [1]byte
}

// Size implements Encodable.
func (e *Uint8) Size() int {
	return 1
}

// Type implements Encodable.
func (e *Uint8) Type() reflect.Type {
	return uint8Type
}

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
func NewUint16() *Uint16 {
	return &Uint16{}
}

// Uint16 is an Encodable for uint16s.
type Uint16 struct {
	buff [2]byte
}

// Size implements Encodable.
func (e *Uint16) Size() int {
	return 2
}

// Type implements Encodable.
func (e Uint16) Type() reflect.Type {
	return uint16Type
}

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
func NewUint32() *Uint32 {
	return &Uint32{}
}

// Uint32 is an Encodable for uint32s.
type Uint32 struct {
	buff [4]byte
}

// Size implements Encodable.
func (e *Uint32) Size() int {
	return 4
}

// Type implements Encodable.
func (e *Uint32) Type() reflect.Type {
	return uint32Type
}

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
func NewUint64() *Uint64 {
	return &Uint64{}
}

// Uint64 is an Encodable for uint64s.
type Uint64 struct {
	buff [8]byte
}

// Size implements Encodable.
func (e *Uint64) Size() int {
	return 8
}

// Type implements Encodable.
func (e *Uint64) Type() reflect.Type {
	return uint64Type
}

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
func NewUint() *Uint {
	return &Uint{}
}

// Uint is an Encodable for uints.
type Uint struct {
	buff [9]byte
}

const (
	maxSingleUint = 255 - 8
)

// Size implements Encodable.
func (e *Uint) Size() int {
	return 9
}

// Type implements Encodable.
func (e *Uint) Type() reflect.Type {
	return uintType
}

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

// NewInt8 returns a new int8 Encodable.
func NewInt8() *Int8 {
	return &Int8{}
}

// Int8 is an Encodable for int8s.
type Int8 struct {
	buff [1]byte
}

// Size implements Encodable.
func (e *Int8) Size() int {
	return 1
}

// Type implements Encodable.
func (e *Int8) Type() reflect.Type {
	return int8Type
}

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
func NewInt16() *Int16 {
	return &Int16{}
}

// Int16 is an Encodable for int16s.
type Int16 struct {
	buff [2]byte
}

// Size implements Encodable.
func (e *Int16) Size() int {
	return 2
}

// Type implements Encodable.
func (e *Int16) Type() reflect.Type {
	return int16Type
}

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
func NewInt32() *Int32 {
	return &Int32{}
}

// Int32 is an Encodable for int32s.
type Int32 struct {
	buff [4]byte
}

// Size implements Encodable.
func (e *Int32) Size() int {
	return 4
}

// Type implements Encodable.
func (e *Int32) Type() reflect.Type {
	return int32Type
}

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
func NewInt64() *Int64 {
	return &Int64{}
}

// Int64 is an Encodable for int64s.
type Int64 struct {
	buff [8]byte
}

// Size implements Encodable.
func (e *Int64) Size() int {
	return 8
}

// Type implements Encodable.
func (e *Int64) Type() reflect.Type {
	return int64Type
}

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
func NewInt() *Int {
	return &Int{}
}

// Int is an Encodable for ints.
type Int struct {
	buff [9]byte
}

const minSingleInt = int8(-1<<7 + 9)

// Size implements Encodable.
func (e *Int) Size() int {
	return 9
}

// Type implements Encodable.
func (e *Int) Type() reflect.Type {
	return intType
}

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

// NewUintptr returns a new uintptr Encodable.
func NewUintptr() *Uintptr {
	return &Uintptr{}
}

// Uintptr is an Encodable for uintptrs.
type Uintptr struct {
	buff [9]byte
}

// Size implements Encodable.
func (e *Uintptr) Size() int {
	return 9
}

// Type implements Encodable.
func (e *Uintptr) Type() reflect.Type {
	return uintptrType
}

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
