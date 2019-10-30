package encodable

import (
	"io"
	"reflect"
	"unsafe"

	"github.com/stewi1014/encs/encio"
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
// Extreme care must be taken, errors from Memory can be difficult to read, let alone helpful in debugging, and are often in the form of panics,
// or worse still, the silent destruction of the universe.
type Memory slicePtr

// String implements Encodable
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
		return encio.Error{
			Err:    encio.ErrNilPointer,
			Caller: "enc.Memory.Encode",
		}
	}
	e.array = ptr
	return encio.Write(*(*[]byte)(unsafe.Pointer(e)), w)
}

// Decode implements Decodable
func (e *Memory) Decode(ptr unsafe.Pointer, r io.Reader) error {
	if ptr == nil {
		return encio.Error{
			Err:    encio.ErrNilPointer,
			Caller: "enc.Memory.Decode",
		}
	}
	e.array = ptr
	return encio.Read(*(*[]byte)(unsafe.Pointer(e)), r)
}
