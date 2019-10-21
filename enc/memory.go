package enc

import (
	"io"
	"reflect"
	"unsafe"
)

// NewMemory returns a new Memory encoder
func NewMemory(size int) *Memory {
	return &Memory{
		len: size,
		cap: size,
	}
}

// Memory is an encoder that throws type-safety out the window (as if the rest of this library wasn't enough).
// Initialised with NewMemory(size), it reads/writes directly to the memory at the given address with no internal buffering.
// Extreme care must be taken, errors from Memory can be difficult to read, let alone helpful in debugging, and are often in the form of unrecoverable panics.
type Memory slicePtr

func (e *Memory) String() string {
	return "Memory"
}

// Type implements Encodable
func (e *Memory) Type() reflect.Type {
	return invalidType
}

// Size implemenets Encodable
func (e *Memory) Size() int {
	return e.len
}

// Encode implements Encodable
func (e *Memory) Encode(ptr unsafe.Pointer, w io.Writer) error {
	if ptr == nil {
		return ErrNilPointer
	}
	e.array = ptr
	return write(*(*[]byte)(unsafe.Pointer(&e)), w)
}

// Decode implements Decodable
func (e *Memory) Decode(ptr unsafe.Pointer, r io.Reader) error {
	if ptr == nil {
		return ErrNilPointer
	}
	e.array = ptr
	return read(*(*[]byte)(unsafe.Pointer(&e)), r)
}
