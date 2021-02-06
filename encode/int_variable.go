package encode

import (
	"fmt"
	"io"
	"reflect"
	"unsafe"

	"github.com/stewi1014/encs/encio"
)

// NewVarint returns a new Varint.
func NewVarint(ty reflect.Type) *Varint {
	v := &Varint{
		ty:   ty,
		buff: make([]byte, 9),
	}

	switch ty.Kind() {
	case reflect.Int:
		v.assignFloat32 = func(n float32, ptr unsafe.Pointer) { *(*int)(ptr) = int(n) }
		v.assignFloat64 = func(n float64, ptr unsafe.Pointer) { *(*int)(ptr) = int(n) }
		v.header = byte(unsafe.Sizeof(int(0))-1) | varOnesFill
	case reflect.Int8:
		v.assignFloat32 = func(n float32, ptr unsafe.Pointer) { *(*int8)(ptr) = int8(n) }
		v.assignFloat64 = func(n float64, ptr unsafe.Pointer) { *(*int8)(ptr) = int8(n) }
		v.header = varOnesFill
	case reflect.Int16:
		v.assignFloat32 = func(n float32, ptr unsafe.Pointer) { *(*int16)(ptr) = int16(n) }
		v.assignFloat64 = func(n float64, ptr unsafe.Pointer) { *(*int16)(ptr) = int16(n) }
		v.header = 1 | varOnesFill
	case reflect.Int32:
		v.assignFloat32 = func(n float32, ptr unsafe.Pointer) { *(*int32)(ptr) = int32(n) }
		v.assignFloat64 = func(n float64, ptr unsafe.Pointer) { *(*int32)(ptr) = int32(n) }
		v.header = 3 | varOnesFill
	case reflect.Int64:
		v.assignFloat32 = func(n float32, ptr unsafe.Pointer) { *(*int64)(ptr) = int64(n) }
		v.assignFloat64 = func(n float64, ptr unsafe.Pointer) { *(*int64)(ptr) = int64(n) }
		v.header = 7 | varOnesFill
	case reflect.Uint:
		v.assignFloat32 = func(n float32, ptr unsafe.Pointer) { *(*uint)(ptr) = uint(n) }
		v.assignFloat64 = func(n float64, ptr unsafe.Pointer) { *(*uint)(ptr) = uint(n) }
		v.header = byte(unsafe.Sizeof(uint(0))) - 1
	case reflect.Uint8:
		v.assignFloat32 = func(n float32, ptr unsafe.Pointer) { *(*uint8)(ptr) = uint8(n) }
		v.assignFloat64 = func(n float64, ptr unsafe.Pointer) { *(*uint8)(ptr) = uint8(n) }
		v.header = 0
	case reflect.Uint16:
		v.assignFloat32 = func(n float32, ptr unsafe.Pointer) { *(*uint16)(ptr) = uint16(n) }
		v.assignFloat64 = func(n float64, ptr unsafe.Pointer) { *(*uint16)(ptr) = uint16(n) }
		v.header = 1
	case reflect.Uint32:
		v.assignFloat32 = func(n float32, ptr unsafe.Pointer) { *(*uint32)(ptr) = uint32(n) }
		v.assignFloat64 = func(n float64, ptr unsafe.Pointer) { *(*uint32)(ptr) = uint32(n) }
		v.header = 3
	case reflect.Uint64:
		v.assignFloat32 = func(n float32, ptr unsafe.Pointer) { *(*uint64)(ptr) = uint64(n) }
		v.assignFloat64 = func(n float64, ptr unsafe.Pointer) { *(*uint64)(ptr) = uint64(n) }
		v.header = 7
	case reflect.Uintptr:
		v.assignFloat32 = func(n float32, ptr unsafe.Pointer) { *(*uintptr)(ptr) = uintptr(n) }
		v.assignFloat64 = func(n float64, ptr unsafe.Pointer) { *(*uintptr)(ptr) = uintptr(n) }
		v.header = byte(unsafe.Sizeof(uintptr(0))) - 1
	case reflect.Float32:
		v.assignInt = func(n int64, ptr unsafe.Pointer) { *(*float32)(ptr) = float32(n) }
		v.assignUint = func(n uint64, ptr unsafe.Pointer) { *(*float32)(ptr) = float32(n) }
		v.assignFloat64 = func(n float64, ptr unsafe.Pointer) { *(*float32)(ptr) = float32(n) }
		v.header = varFloat32
	case reflect.Float64:
		v.assignInt = func(n int64, ptr unsafe.Pointer) { *(*float64)(ptr) = float64(n) }
		v.assignUint = func(n uint64, ptr unsafe.Pointer) { *(*float64)(ptr) = float64(n) }
		v.assignFloat32 = func(n float32, ptr unsafe.Pointer) { *(*float64)(ptr) = float64(n) }
		v.header = varFloat64
	default:
		panic(encio.NewError(encio.ErrBadType, fmt.Sprintf("%v is not an integer or float kind", ty.String()), 0))
	}

	return v
}

// Varint is an Encodable for integer and float types.
// It can decode from a Variant initialised with int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint32, uint64, uintptr, float32 and float64.
// It can only encode from the type is it initialised with.
type Varint struct {
	ty            reflect.Type
	buff          []byte
	bits          uint64
	header        byte
	assignFloat32 func(float32, unsafe.Pointer)
	assignFloat64 func(float64, unsafe.Pointer)
	assignUint    func(uint64, unsafe.Pointer)
	assignInt     func(int64, unsafe.Pointer)
}

const (
	varSizeMask byte = 1<<3 - 1
	varFloat    byte = 1 << (iota + 3)
	varOnesFill

	varFloat32 = varFloat | 3
	varFloat64 = varFloat | 7
)

// Size implements Encodable.
func (e *Varint) Size() int { return int(e.header&varSizeMask) + 2 }

// Type implements Encodable.
func (e *Varint) Type() reflect.Type { return e.ty }

// Encode implements Encodable.
func (e *Varint) Encode(ptr unsafe.Pointer, w io.Writer) error {
	e.buff[0] = e.header &^ varOnesFill
	switch e.header & varSizeMask {
	case 0:
		e.buff[1] = *(*uint8)(ptr)
	case 1:
		i := *(*uint16)(ptr)
		e.buff[1] = uint8(i)
		e.buff[2] = uint8(i >> 8)
	case 3:
		i := *(*uint32)(ptr)
		e.buff[1] = uint8(i)
		e.buff[2] = uint8(i >> 8)
		e.buff[3] = uint8(i >> 16)
		e.buff[4] = uint8(i >> 24)
	case 7:
		i := *(*uint64)(ptr)
		e.buff[1] = uint8(i)
		e.buff[2] = uint8(i >> 8)
		e.buff[3] = uint8(i >> 16)
		e.buff[4] = uint8(i >> 24)
		e.buff[5] = uint8(i >> 32)
		e.buff[6] = uint8(i >> 40)
		e.buff[7] = uint8(i >> 48)
		e.buff[8] = uint8(i >> 56)
	default:
		panic("impossible varint header. must be radiation flipping bits.")
	}

	// compress
	if e.header&varFloat == 0 {
		var fill byte
		if e.buff[e.buff[0]&varSizeMask+1]&(1<<7) != 0 && e.header&varOnesFill != 0 {
			fill = 1<<8 - 1
			e.buff[0] |= varOnesFill
		}

		for i := e.buff[0]&varSizeMask + 1; e.buff[i] == fill && fill&(1<<7)&^e.buff[i-1]&(1<<7) == 0 && i >= 2; i-- {
			e.buff[0]--
		}
	}

	return encio.Write(e.buff[:(e.buff[0]&varSizeMask)+2], w)
}

// Decode implements Encodable.
func (e *Varint) Decode(ptr unsafe.Pointer, r io.Reader) error {
	if err := encio.Read(e.buff[:1], r); err != nil {
		return err
	}

	if err := encio.Read(e.buff[1:e.buff[0]&varSizeMask+2], r); err != nil {
		return err
	}

	// decompress
	shift := int8(e.header&varSizeMask) - int8(e.buff[0]&varSizeMask)
	dptr := ptr
	if e.buff[0]&varFloat == 0 {
	doshift:
		switch {
		case shift > 0:
			var fill byte
			if e.buff[0]&varOnesFill != 0 {
				fill = 1<<8 - 1
			}

			for i := e.buff[0]&varSizeMask + 2; i < (byte(shift) + e.buff[0]&varSizeMask + 2); i++ {
				e.buff[i] = fill
			}
			e.buff[0] += byte(shift)
		case shift < 0 && e.header&varFloat == 0:
			e.buff[0] += byte(shift) // actually a subtraction.
		case shift < 0 && e.header&varFloat != 0:
			dptr = unsafe.Pointer(&e.bits)
			shift = 7 - int8(e.buff[0]&varSizeMask)
			goto doshift
		}
	} else if shift < 0 {
		// must be float -> smaller number
		dptr = unsafe.Pointer(&e.bits)
	}

	switch e.buff[0] & (varSizeMask | varFloat) {
	case 0:
		*(*uint8)(dptr) = e.buff[1]
	case 1:
		i := (*uint16)(dptr)
		*i = uint16(e.buff[1])
		*i |= uint16(e.buff[2]) << 8
	case 3, varFloat32:
		i := (*uint32)(dptr)
		*i = uint32(e.buff[1])
		*i |= uint32(e.buff[2]) << 8
		*i |= uint32(e.buff[3]) << 16
		*i |= uint32(e.buff[4]) << 24
	case 7, varFloat64:
		i := (*uint64)(dptr)
		*i = uint64(e.buff[1])
		*i |= uint64(e.buff[2]) << 8
		*i |= uint64(e.buff[3]) << 16
		*i |= uint64(e.buff[4]) << 24
		*i |= uint64(e.buff[5]) << 32
		*i |= uint64(e.buff[6]) << 40
		*i |= uint64(e.buff[7]) << 48
		*i |= uint64(e.buff[8]) << 56
	default:
		// If it's an int type, we should have already filled or removed to make it a proper size.
		// If it's a float type it can only be 3(4) or 7(8).
		return encio.NewError(encio.ErrMalformed, fmt.Sprintf("impossible varint header %8b", e.buff[0]), 0)
	}

	if e.buff[0] != e.header {
		// Headers don't match.

		if e.buff[0]&varFloat == 0 && e.header&varFloat == 0 {
			// Both are integers, just different sizes.
			// This has already been handled
			return nil
		}
		switch e.buff[0] & (varSizeMask | varFloat | varOnesFill) {
		case varFloat32:
			e.assignFloat32(*(*float32)(dptr), ptr)
		case varFloat64:
			e.assignFloat64(*(*float64)(dptr), ptr)

		// varOnesFill | 0, varOneFill | 1 are both filled away if we're decoding into a float type.
		// assignment not needed if this is an int type.
		case varOnesFill | 3:
			e.assignInt(int64(*(*int32)(dptr)), ptr)
		case varOnesFill | 7:
			e.assignInt(int64(*(*int64)(dptr)), ptr)

		// 0, 1 are both filled away if we're decoding into a float type.
		case 3:
			e.assignUint(uint64(*(*uint32)(dptr)), ptr)
		case 7:
			e.assignUint(uint64(*(*uint64)(dptr)), ptr)
		default:
			return encio.NewError(encio.ErrMalformed, fmt.Sprintf("cannot assign from number type with header %8b", e.buff[0]), 0)
		}
	}

	return nil
}
