package enc

import (
	"io"
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
type Memory slicePtr

// Size implemenets Sized
func (m *Memory) Size() int {
	return m.len
}

// Encode implements Encodable
func (m *Memory) Encode(ptr unsafe.Pointer, w io.Writer) error {
	if ptr == nil {
		return ErrNilPointer
	}
	m.array = ptr
	return write(*(*[]byte)(unsafe.Pointer(&m)), w)
}

// Decode implements Decodable
func (m *Memory) Decode(ptr unsafe.Pointer, r io.Reader) error {
	if ptr == nil {
		return ErrNilPointer
	}
	m.array = ptr
	return read(*(*[]byte)(unsafe.Pointer(&m)), r)
}
