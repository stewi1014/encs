// Package enc provides low-level methods for seralising golang data structures.
// It aims to be fast, modular and comprehensive, valuing runtime speed over creation overhead.
//
// Encodable is an encoder/decoder for a specific type.
//
// The pointers passed to Encode and Decode *must* be pointers to an allocated instance of the Encodable's type, accessible by Type().
// If not, expect arbritrary memory addresses to be modified, and unhelpful runtime errors. Reference types will point themselves to newly allocated types,
// but the reference type itself must be allocated beforehand.
//
// Encodables typically contain static buffers, used by calls to Encode and Decode. Concurrent usage will certainly fail.
// If a concurrent-safe Encodable is needed, NewConcurrent is a drop-in, concurrent safe replacement.
//
// NewEncodable creates an Encodable for a type.
// It takes a *Config, which is used to configure the Encodable.
//
//
// Type is an encoder/decoder for type information; an encoder for reflect.Type.
// The primary implementation, RegisterType
package enc

import (
	"errors"
	"fmt"
	"io"
	"reflect"
	"unsafe"
)

var (
	// TooBig is a byte count used for simple sanity checking before things like allocation and iteration with numbers decoded from readers.
	// By default it is 32MB on 32bit machines, and 128MB on 64bit machines.
	// Feel free to change it.
	TooBig = 1 << (25 + ((^uint(0) >> 32) & 2))
)

// Error constants.
// Errors should be checked with errors.Is() or errors.As().
// Returned errors can wrap multiple errors; i.e. an unexpected EOF will wrap a ErrMalformed and io.ErrUnexpectedEOF.
var (
	// ErrMalformed is returned if the read data is not valid for decoding,
	// including read errors and short reads.
	ErrMalformed = errors.New("malformed")

	// ErrBadType is returned when a type, where possible to detect, is wrong, unresolvable or inappropriate.
	// Due to the usage of unsafe.Pointer, it is not usually possible to detect incorrect types.
	// If this error is seen, it should be taken seriously; encoding of incorrect types has undefined behaviour.
	ErrBadType = errors.New("bad type")

	// ErrNilPointer is returned if an encoder or decoder has a nil pointer it can't resolve.
	ErrNilPointer = errors.New("nil pointer")
)

// Encodable is an Encoder/Decoder for a specific type.
// Encode() and Decode() often share static buffers for the sake of performance, but this comes at the cost of thread safety.
// Concurrent calls to either Encode() or Decode() will almost certainly result in complete failure.
// Use NewConcurrent if concurrency is needed.
type Encodable interface {
	// Type returns the type that the Encodable encodes.
	// ptr in Encode() and Decode() *must* be a pointer to an object of this type.
	// It is thread safe.
	Type() reflect.Type

	// Encode writes the encoded form of the object at ptr to w.
	Encode(ptr unsafe.Pointer, w io.Writer) error

	// Decode reads the encoded form from r into the object at ptr.
	// Decodes will only read what Encode wrote.
	Decode(ptr unsafe.Pointer, r io.Reader) error

	// Size returns the maximum encoded size of the Encodable.
	// If Size returns <0, size is undefined.
	// It is thread safe.
	Size() int

	// String returns a string showing the Encodable's structure.
	// () shows a little information about the encoder, such as the type it encodes or relavent settings.
	// {} denotes sub-Encoders; encoders that encode parts of a larger Encodable.
	// Comparing results from String is an effecive way of equality checking.
	// It is thread safe.
	fmt.Stringer // String() string
}

// NewEncodable returns a new Encodable for encoding the type t.
// config contains settings and information for the generation of the Encodable.
// In many cases, it can be nil for sane defaults, however some Enodable types require information from the config.
func NewEncodable(t reflect.Type, config *Config) Encodable {
	// copy config; changes can be made to it and we want to be able to re-use it without stepping on other Encodables.
	if config != nil {
		config = config.copy()
	}
	return newEncodable(t, config)
}

func newEncodable(t reflect.Type, config *Config) Encodable {
	if config == nil {
		config = new(Config)
	}

	ptrt := reflect.PtrTo(t)
	kind := t.Kind()
	switch {
	// Implementers
	case ptrt.Implements(binaryMarshalerIface) && ptrt.Implements(binaryUnmarshalerIface):
		return NewBinaryMarshaler(t)

	// Compound-Types
	case kind == reflect.Ptr:
		return newPointer(t, config)
	case kind == reflect.Interface:
		return newInterface(t, config)
	case kind == reflect.Struct:
		return newStruct(t, config)
	case kind == reflect.Array:
		return newArray(t, config)
	case kind == reflect.Slice:
		return newSlice(t, config)
	case kind == reflect.Map:
		return newMap(t, config)

	// Integer types
	case kind == reflect.Uint8:
		return NewUint8()
	case kind == reflect.Uint16:
		return NewUint16()
	case kind == reflect.Uint32:
		return NewUint32()
	case kind == reflect.Uint64:
		return NewUint64()
	case kind == reflect.Uint:
		return NewUint()
	case kind == reflect.Int8:
		return NewInt8()
	case kind == reflect.Int16:
		return NewInt16()
	case kind == reflect.Int32:
		return NewInt32()
	case kind == reflect.Int64:
		return NewInt64()
	case kind == reflect.Int:
		return NewInt()
	case kind == reflect.Uintptr:
		return NewUintptr()

	// Float types
	case kind == reflect.Float32:
		return NewFloat32()
	case kind == reflect.Float64:
		return NewFloat64()
	case kind == reflect.Complex64:
		return NewComplex64()
	case kind == reflect.Complex128:
		return NewComplex128()

	// Misc types
	case kind == reflect.Bool:
		return NewBool()
	case kind == reflect.String:
		return NewString()
	}

	panic(fmt.Errorf("%v: cannot find encoder for type %v", ErrBadType, t))
}
