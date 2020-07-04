// Package encodable provides low-level methods for seralising golang data structures.
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
package encodable

// I intend to keep a curated lsit of important notes to keep in mind while developing this part of encs here.
//
// Every instance of unsafe.Pointer that exists must always point towards a valid object
// unsafe.Poointer types are functional pointers with all the semantics that come with it;
// at any time, the garbage collector could come along, try to dereference the pointer and crash the program. yikes.
// If an invalid pointer is needed as an intermediary step, uintptr should be used.

import (
	"fmt"
	"io"
	"reflect"
	"unsafe"

	"github.com/stewi1014/encs/encio"
)

// Encodable is an Encoder/Decoder for a specific type.
//
// Encoders are not assumed to be thread safe.
// Encode() and Decode() often share static buffers for the sake of performance, but this comes at the cost of thread safety.
// Concurrent calls to either Encode() or Decode() will almost certainly result in complete failure.
// Use NewConcurrent if concurrency is needed.
//
// Encodables should return two kinds of error.
// encio.IOError and encio.Error for io and corrupted data errors, and Error for encoding errors.
// See encs/encio
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
	// () shows a little information about the encoder, such as the type it encodes or relevant settings.
	// {} denotes sub-Encoders; encoders that encode parts of a larger Encodable.
	// Comparing results from String is an effecive way of equality checking.
	// It is thread safe.
	fmt.Stringer // String() string
}

// New returns a new Encodable for encoding the type t.
// config contains settings and information for the generation of the Encodable.
// In many cases, it can be nil for sane defaults, however some Enodable types require information from the config.
func New(t reflect.Type, config *Config) Encodable {
	return newEncodable(t, config.genState())
}

// newEncodable creates a new Encodable from state.
// as a general rule, New* functions are for creating new, independent Encodables,
// while new* functions are for creating encodables that are children of existing Encodables.
func newEncodable(t reflect.Type, state *state) Encodable {
	ptrt := reflect.PtrTo(t)
	kind := t.Kind()
	switch {
	// Implementers
	case ptrt.Implements(binaryMarshalerIface) && ptrt.Implements(binaryUnmarshalerIface):
		return NewBinaryMarshaler(t)

	// Compound-Types
	case kind == reflect.Ptr:
		return newPointer(t, state)
	case kind == reflect.Interface:
		return newInterface(t, state)
	case kind == reflect.Struct:
		return newStruct(t, state)
	case kind == reflect.Array:
		return newArray(t, state)
	case kind == reflect.Slice:
		return newSlice(t, state)
	case kind == reflect.Map:
		return newMap(t, state)

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

	panic(encio.NewError(encio.ErrBadType, fmt.Sprintf("cannot create encodable for type %v", t), 0))
}

// NewSource returns a Source with the given config and new function.
func NewSource(config *Config, newFunc func(reflect.Type, *Config) Encodable) *Source {
	// we must hold config, so we copy it
	config = config.copy()

	return &Source{
		encs:   make(map[reflect.Type]Encodable),
		config: config,
		new:    newFunc,
	}
}

// Source is a cache of Encodables. Encodables are only created once, with subsequent calls to GetEncodable returning the previously created encodable.
type Source struct {
	encs   map[reflect.Type]Encodable
	config *Config
	new    func(reflect.Type, *Config) Encodable
}

// GetEncodable returns an encodable for the given type, as created by the new function passed to NewSource.
func (s *Source) GetEncodable(ty reflect.Type) Encodable {
	if enc, ok := s.encs[ty]; ok {
		return enc
	}

	enc := s.new(ty, s.config)
	s.encs[ty] = enc
	return enc
}
