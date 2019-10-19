package enc

// Float & Complex type encoders

import (
	"io"
	"reflect"
	"unsafe"
)

// NewFloat32 returns a new float32 Encodable.
func NewFloat32() *Float32 {
	return &Float32{
		buff: make([]byte, 4),
	}
}

// Float32 is an Encodable for float32s
type Float32 struct {
	buff []byte
}

// Size implemenets Sized
func (e *Float32) Size() int {
	return 4
}

// Type implements Encodable
func (e *Float32) Type() reflect.Type {
	return float32Type
}

// Encode implements Encodable
func (e *Float32) Encode(ptr unsafe.Pointer, w io.Writer) error {
	if ptr == nil {
		return ErrNilPointer
	}
	bits := *(*uint32)(ptr)
	e.buff[0] = uint8(bits)
	e.buff[1] = uint8(bits >> 8)
	e.buff[2] = uint8(bits >> 16)
	e.buff[3] = uint8(bits >> 24)

	return write(e.buff, w)
}

// Decode implements Encodable
func (e *Float32) Decode(ptr unsafe.Pointer, r io.Reader) error {
	if ptr == nil {
		return ErrNilPointer
	}
	if err := read(e.buff, r); err != nil {
		return err
	}
	bits := (*uint32)(ptr)
	*bits = uint32(e.buff[0])
	*bits |= uint32(e.buff[1]) << 8
	*bits |= uint32(e.buff[2]) << 16
	*bits |= uint32(e.buff[3]) << 24

	return nil
}

// NewFloat64 returns a new float64 Encodable.
func NewFloat64() *Float64 {
	return &Float64{
		buff: make([]byte, 8),
	}
}

// Float64 is an Encodable for float64s
type Float64 struct {
	buff []byte
}

// Size implemenets Sized
func (e *Float64) Size() int {
	return 8
}

// Type implements Encodable
func (e *Float64) Type() reflect.Type {
	return float64Type
}

// Encode implements Encodable
func (e *Float64) Encode(ptr unsafe.Pointer, w io.Writer) error {
	if ptr == nil {
		return ErrNilPointer
	}
	bits := *(*uint64)(ptr)
	e.buff[0] = uint8(bits)
	e.buff[1] = uint8(bits >> 8)
	e.buff[2] = uint8(bits >> 16)
	e.buff[3] = uint8(bits >> 24)
	e.buff[4] = uint8(bits >> 32)
	e.buff[5] = uint8(bits >> 40)
	e.buff[6] = uint8(bits >> 48)
	e.buff[7] = uint8(bits >> 56)

	return write(e.buff, w)
}

// Decode implements Encodable
func (e *Float64) Decode(ptr unsafe.Pointer, r io.Reader) error {
	if ptr == nil {
		return ErrNilPointer
	}
	if err := read(e.buff, r); err != nil {
		return err
	}
	bits := (*uint64)(ptr)
	*bits = uint64(e.buff[0])
	*bits |= uint64(e.buff[1]) << 8
	*bits |= uint64(e.buff[2]) << 16
	*bits |= uint64(e.buff[3]) << 24
	*bits |= uint64(e.buff[4]) << 32
	*bits |= uint64(e.buff[5]) << 40
	*bits |= uint64(e.buff[6]) << 48
	*bits |= uint64(e.buff[7]) << 56

	return nil
}

// NewComplex64 returns a new complex128 Encodable
func NewComplex64() *Complex64 {
	return &Complex64{
		buff: make([]byte, 8),
	}
}

// Complex64 is an Encodable for complex64s
type Complex64 struct {
	buff []byte
}

// Size implemenets Sized
func (e *Complex64) Size() int {
	return 8
}

// Type implements Encodable
func (e *Complex64) Type() reflect.Type {
	return complex64Type
}

// Encode implements Encodable
func (e *Complex64) Encode(ptr unsafe.Pointer, w io.Writer) error {
	if ptr == nil {
		return ErrNilPointer
	}
	bits := *(*uint32)(ptr)
	e.buff[0] = uint8(bits)
	e.buff[1] = uint8(bits >> 8)
	e.buff[2] = uint8(bits >> 16)
	e.buff[3] = uint8(bits >> 24)

	bits = *(*uint32)(unsafe.Pointer(uintptr(ptr) + 4))
	e.buff[4] = uint8(bits)
	e.buff[5] = uint8(bits >> 8)
	e.buff[6] = uint8(bits >> 16)
	e.buff[7] = uint8(bits >> 24)

	return write(e.buff, w)
}

// Decode implements Encodable
func (e *Complex64) Decode(ptr unsafe.Pointer, r io.Reader) error {
	if ptr == nil {
		return ErrNilPointer
	}
	if err := read(e.buff, r); err != nil {
		return err
	}
	bits := (*uint32)(ptr)
	*bits = uint32(e.buff[0])
	*bits |= uint32(e.buff[1]) << 8
	*bits |= uint32(e.buff[2]) << 16
	*bits |= uint32(e.buff[3]) << 24

	bits = (*uint32)(unsafe.Pointer(uintptr(ptr) + 4))
	*bits = uint32(e.buff[4])
	*bits |= uint32(e.buff[5]) << 8
	*bits |= uint32(e.buff[6]) << 16
	*bits |= uint32(e.buff[7]) << 24

	return nil
}

// NewComplex128 returns a new complex128 Encodable
func NewComplex128() *Complex128 {
	return &Complex128{
		buff: make([]byte, 16),
	}
}

// Complex128 is an Encodable for complex128s
type Complex128 struct {
	buff []byte
}

// Size implemenets Sized
func (e *Complex128) Size() int {
	return 16
}

// Type implements Encodable
func (e *Complex128) Type() reflect.Type {
	return complex128Type
}

// Encode implements Encodable
func (e *Complex128) Encode(ptr unsafe.Pointer, w io.Writer) error {
	if ptr == nil {
		return ErrNilPointer
	}
	bits := *(*uint64)(ptr)
	e.buff[0] = uint8(bits)
	e.buff[1] = uint8(bits >> 8)
	e.buff[2] = uint8(bits >> 16)
	e.buff[3] = uint8(bits >> 24)
	e.buff[4] = uint8(bits >> 32)
	e.buff[5] = uint8(bits >> 40)
	e.buff[6] = uint8(bits >> 48)
	e.buff[7] = uint8(bits >> 56)

	bits = *(*uint64)(unsafe.Pointer(uintptr(ptr) + 8))
	e.buff[8] = uint8(bits)
	e.buff[9] = uint8(bits >> 8)
	e.buff[10] = uint8(bits >> 16)
	e.buff[11] = uint8(bits >> 24)
	e.buff[12] = uint8(bits >> 32)
	e.buff[13] = uint8(bits >> 40)
	e.buff[14] = uint8(bits >> 48)
	e.buff[15] = uint8(bits >> 56)

	return write(e.buff, w)
}

// Decode implements Encodable
func (e *Complex128) Decode(ptr unsafe.Pointer, r io.Reader) error {
	if ptr == nil {
		return ErrNilPointer
	}
	if err := read(e.buff, r); err != nil {
		return err
	}
	bits := (*uint64)(ptr)
	*bits = uint64(e.buff[0])
	*bits |= uint64(e.buff[1]) << 8
	*bits |= uint64(e.buff[2]) << 16
	*bits |= uint64(e.buff[3]) << 24
	*bits |= uint64(e.buff[4]) << 32
	*bits |= uint64(e.buff[5]) << 40
	*bits |= uint64(e.buff[6]) << 48
	*bits |= uint64(e.buff[7]) << 56

	bits = (*uint64)(unsafe.Pointer(uintptr(ptr) + 8))
	*bits = uint64(e.buff[8])
	*bits |= uint64(e.buff[9]) << 8
	*bits |= uint64(e.buff[10]) << 16
	*bits |= uint64(e.buff[11]) << 24
	*bits |= uint64(e.buff[12]) << 32
	*bits |= uint64(e.buff[13]) << 40
	*bits |= uint64(e.buff[14]) << 48
	*bits |= uint64(e.buff[15]) << 56

	return nil
}
