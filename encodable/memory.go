package encodable

import (
	"io"
	"reflect"
	"unsafe"

	"github.com/stewi1014/encs/encio"
)

// NewMemory returns a new Memory encoder.
func NewMemory(size int) *Memory {
	return &Memory{
		buff: reflect.SliceHeader{
			Len: size,
			Cap: size,
		},
	}
}

// Memory is an encoder that throws type-safety out the window.
// Initialised with NewMemory(size), it reads/writes directly to the memory at the given address with no internal buffering.
// Extreme care must be taken, errors from Memory can be difficult to read, let alone helpful in debugging, and are often in the form of panics,
// or worse still, the silent destruction of the universe.
type Memory struct {
	buff reflect.SliceHeader
}

// Type implements Encodable.
func (e *Memory) Type() reflect.Type {
	return nil
}

// Size implemenets Encodable.
func (e *Memory) Size() int {
	return e.buff.Cap
}

// Encode implements Encodable.
func (e *Memory) Encode(ptr unsafe.Pointer, w io.Writer) error {
	checkPtr(ptr)
	e.buff.Data = uintptr(ptr)
	return encio.Write(*(*[]byte)(unsafe.Pointer(&e.buff)), w)
}

// Decode implements Decodable.
func (e *Memory) Decode(ptr unsafe.Pointer, r io.Reader) error {
	checkPtr(ptr)
	e.buff.Data = uintptr(ptr)
	return encio.Read(*(*[]byte)(unsafe.Pointer(&e.buff)), r)
}
