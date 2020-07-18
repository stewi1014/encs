package encodable

// Float & Complex type encoders

import (
	"fmt"
	"io"
	"reflect"
	"unsafe"

	"github.com/stewi1014/encs/encio"
)

// NewFloat32 returns a new float32 Encodable.
func NewFloat32(ty reflect.Type) *Float32 {
	if ty.Kind() != reflect.Float32 {
		panic(encio.NewError(encio.ErrBadType, fmt.Sprintf("%v is not of float32 kind", ty.String()), 0))
	}
	return &Float32{
		ty:   ty,
		buff: make([]byte, 4),
	}
}

// Float32 is an Encodable for float32s.
type Float32 struct {
	ty   reflect.Type
	buff []byte
}

// Size implemenets Encodable.
func (e *Float32) Size() int { return 4 }

// Type implements Encodable.
func (e *Float32) Type() reflect.Type { return e.ty }

// Encode implements Encodable.
func (e *Float32) Encode(ptr unsafe.Pointer, w io.Writer) error {
	checkPtr(ptr)
	bits := *(*uint32)(ptr)
	e.buff[0] = uint8(bits)
	e.buff[1] = uint8(bits >> 8)
	e.buff[2] = uint8(bits >> 16)
	e.buff[3] = uint8(bits >> 24)

	return encio.Write(e.buff, w)
}

// Decode implements Encodable.
func (e *Float32) Decode(ptr unsafe.Pointer, r io.Reader) error {
	checkPtr(ptr)
	if err := encio.Read(e.buff, r); err != nil {
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
func NewFloat64(ty reflect.Type) *Float64 {
	if ty.Kind() != reflect.Float64 {
		panic(encio.NewError(encio.ErrBadType, fmt.Sprintf("%v is not of float64 kind", ty.String()), 0))
	}
	return &Float64{
		ty:   ty,
		buff: make([]byte, 8),
	}
}

// Float64 is an Encodable for float64s.
type Float64 struct {
	ty   reflect.Type
	buff []byte
}

// Size implemenets Encodable.
func (e *Float64) Size() int { return 8 }

// Type implements Encodable.
func (e *Float64) Type() reflect.Type { return e.ty }

// Encode implements Encodable.
func (e *Float64) Encode(ptr unsafe.Pointer, w io.Writer) error {
	checkPtr(ptr)
	bits := *(*uint64)(ptr)
	e.buff[0] = uint8(bits)
	e.buff[1] = uint8(bits >> 8)
	e.buff[2] = uint8(bits >> 16)
	e.buff[3] = uint8(bits >> 24)
	e.buff[4] = uint8(bits >> 32)
	e.buff[5] = uint8(bits >> 40)
	e.buff[6] = uint8(bits >> 48)
	e.buff[7] = uint8(bits >> 56)

	return encio.Write(e.buff, w)
}

// Decode implements Encodable.
func (e *Float64) Decode(ptr unsafe.Pointer, r io.Reader) error {
	checkPtr(ptr)
	if err := encio.Read(e.buff, r); err != nil {
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

// NewComplex returns an Encodable for the given complex type.
// It supports complex64 and complex128.
// If LooseTyping is enabled, it returns VarComplex, which allows encoding between different complex types.
func NewComplex(ty reflect.Type, config Config) Encodable {
	switch {
	case config&LooseTyping != 0:
		return NewVarComplex(ty)
	case ty.Kind() == reflect.Complex128:
		return NewComplex128(ty)
	case ty.Kind() == reflect.Complex64:
		return NewComplex64(ty)
	default:
		panic(encio.NewError(encio.ErrBadType, fmt.Sprintf("%v is not of complex64 or complex128 kind", ty.String()), 0))
	}
}

// NewVarComplex returns a new VarComplex Encodable.
// It supports complex64 and complex128 types.
func NewVarComplex(ty reflect.Type) *VarComplex {
	if ty.Kind() != reflect.Complex64 && ty.Kind() != reflect.Complex128 {
		panic(encio.NewError(encio.ErrBadType, fmt.Sprintf("%v is not a complex64 or complex128 type", ty.String()), 0))
	}

	return &VarComplex{
		buff: make([]byte, 17),
		ty:   ty,
	}
}

// VarComplex is an Encodable for complex64 and complex128 types.
// It can encode to/decode from either, but can only encode from the type given during initialisation.
type VarComplex struct {
	ty   reflect.Type
	buff []byte
}

const (
	varComplex64 = iota
	varComplex128
)

// Size implements Encodable.
func (e *VarComplex) Size() int {
	if e.ty.Kind() == reflect.Complex128 {
		return 17
	}
	return 9
}

// Type implements Encodable.
func (e *VarComplex) Type() reflect.Type { return e.ty }

// Encode implements Encodable.
func (e *VarComplex) Encode(ptr unsafe.Pointer, w io.Writer) error {
	checkPtr(ptr)
	if e.ty.Kind() == reflect.Complex64 {
		e.buff[0] = varComplex64

		checkPtr(ptr)
		bits := *(*uint32)(ptr)
		e.buff[1] = uint8(bits)
		e.buff[2] = uint8(bits >> 8)
		e.buff[3] = uint8(bits >> 16)
		e.buff[4] = uint8(bits >> 24)

		bits = *(*uint32)(unsafe.Pointer(uintptr(ptr) + 4))
		e.buff[5] = uint8(bits)
		e.buff[6] = uint8(bits >> 8)
		e.buff[7] = uint8(bits >> 16)
		e.buff[8] = uint8(bits >> 24)

		return encio.Write(e.buff[:9], w)
	}
	e.buff[0] = varComplex128

	checkPtr(ptr)
	bits := *(*uint64)(ptr)
	e.buff[1] = uint8(bits)
	e.buff[2] = uint8(bits >> 8)
	e.buff[3] = uint8(bits >> 16)
	e.buff[4] = uint8(bits >> 24)
	e.buff[5] = uint8(bits >> 32)
	e.buff[6] = uint8(bits >> 40)
	e.buff[7] = uint8(bits >> 48)
	e.buff[8] = uint8(bits >> 56)

	bits = *(*uint64)(unsafe.Pointer(uintptr(ptr) + 8))
	e.buff[9] = uint8(bits)
	e.buff[10] = uint8(bits >> 8)
	e.buff[11] = uint8(bits >> 16)
	e.buff[12] = uint8(bits >> 24)
	e.buff[13] = uint8(bits >> 32)
	e.buff[14] = uint8(bits >> 40)
	e.buff[15] = uint8(bits >> 48)
	e.buff[16] = uint8(bits >> 56)

	return encio.Write(e.buff, w)
}

// Decode implements Encodable.
func (e *VarComplex) Decode(ptr unsafe.Pointer, r io.Reader) error {
	checkPtr(ptr)
	if err := encio.Read(e.buff[:1], r); err != nil {
		return err
	}

	switch {
	case e.buff[0] == varComplex128 && e.ty.Kind() == reflect.Complex128:
		return e.decode128(ptr, r)

	case e.buff[0] == varComplex64 && e.ty.Kind() == reflect.Complex64:
		return e.decode64(ptr, r)

	case e.buff[0] == varComplex128 && e.ty.Kind() == reflect.Complex64:
		var tmp complex128
		if err := e.decode128(unsafe.Pointer(&tmp), r); err != nil {
			return err
		}

		*(*complex64)(ptr) = complex64(tmp)
		return nil

	case e.buff[0] == varComplex64 && e.ty.Kind() == reflect.Complex128:
		var tmp complex64
		if err := e.decode64(unsafe.Pointer(&tmp), r); err != nil {
			return err
		}

		*(*complex128)(ptr) = complex128(tmp)
		return nil

	default:
		return encio.NewError(encio.ErrMalformed, fmt.Sprintf("variable size complex header doesn't match any known sizes. got %8b", e.buff[0]), 0)
	}
}

func (e *VarComplex) decode64(ptr unsafe.Pointer, r io.Reader) error {
	if err := encio.Read(e.buff[:8], r); err != nil {
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

func (e *VarComplex) decode128(ptr unsafe.Pointer, r io.Reader) error {
	if err := encio.Read(e.buff[:16], r); err != nil {
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

// NewComplex64 returns a new complex128 Encodable.
func NewComplex64(ty reflect.Type) *Complex64 {
	return &Complex64{
		ty:   ty,
		buff: make([]byte, 8),
	}
}

// Complex64 is an Encodable for complex64s.
type Complex64 struct {
	ty   reflect.Type
	buff []byte
}

// Size implemenets Encodable.
func (e *Complex64) Size() int { return 8 }

// Type implements Encodable.
func (e *Complex64) Type() reflect.Type { return e.ty }

// Encode implements Encodable.
func (e *Complex64) Encode(ptr unsafe.Pointer, w io.Writer) error {
	checkPtr(ptr)
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

	return encio.Write(e.buff, w)
}

// Decode implements Encodable.
func (e *Complex64) Decode(ptr unsafe.Pointer, r io.Reader) error {
	checkPtr(ptr)
	if err := encio.Read(e.buff, r); err != nil {
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

// NewComplex128 returns a new complex128 Encodable.
func NewComplex128(ty reflect.Type) *Complex128 {
	if ty.Kind() != reflect.Complex128 {
		panic(encio.NewError(encio.ErrBadType, fmt.Sprintf("%v is not of complex128 kind", ty.String()), 0))
	}

	return &Complex128{
		ty:   ty,
		buff: make([]byte, 16),
	}
}

// Complex128 is an Encodable for complex128s.
type Complex128 struct {
	ty   reflect.Type
	buff []byte
}

// Size implemenets Encodable.
func (e *Complex128) Size() int { return 16 }

// Type implements Encodable.
func (e *Complex128) Type() reflect.Type { return e.ty }

// Encode implements Encodable.
func (e *Complex128) Encode(ptr unsafe.Pointer, w io.Writer) error {
	checkPtr(ptr)
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

	return encio.Write(e.buff, w)
}

// Decode implements Encodable.
func (e *Complex128) Decode(ptr unsafe.Pointer, r io.Reader) error {
	checkPtr(ptr)
	if err := encio.Read(e.buff, r); err != nil {
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
