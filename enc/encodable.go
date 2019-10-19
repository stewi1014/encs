// Package enc provides low-level methods for seralising golang data structures.
// It aims to be fast, modular and comprehensive, valuing runtime speed over creation overhead.
//
// Encodable is an encoder/decoder for a specific type.
// The pointers passed to Encode and Decode *must* be pointers to an allocated instance of the Encodable's type, accessible by Type().
// If not, expect arbritrary memory addresses to be modified, and unhelpful runtime errors. Reference types can point themselves to newly allocated types,
// but the reference type itself must be allocated beforehand.
//
// Encodables typically contain static buffers, used by calls to Encode and Decode. Concurrent usage will certainly fail.
// If a concurrent-safe Encodable is needed, NewConcurrentEncodable is a drop-in, concurrent safe replacement.
//
// NewEncodable creates an Encodable for a type.
// It takes a *Config, which is used to configure the Encodable.
//
// Type is an encoder/decoder for type information; an encoder for reflect.Type.
// The primary implementation, RegisterType
//
// The usage of Encodables is not too different from
//
//
// Encoders and Decoders may read and write varying amounts, but they will always read the same amount that was written.
package enc

import (
	"errors"
	"fmt"
	"io"
	"reflect"
	"unsafe"
)

const (
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
	ErrMalformed = errors.New("malformed buffer")

	// ErrBadType is returned when the given value, where possible to detect, is the wrong type or otherwise innapropriate.
	// Due to the usage of unsafe.Pointer, it is not usually possible to detect incorrect types.
	// If this error is seen, it should be taken seriously; encoding of incorrect types has undefined behaviour.
	ErrBadType = errors.New("bad type")

	// ErrNilPointer is returned if an encoder or decoder has a nil pointer it can't resolve.
	ErrNilPointer = errors.New("nil pointer")
)

// Config contains settings and information for the generation of a new Encodable.
// Some Encodables do nothing with Config, and some require certain elements to be set.
type Config struct {
	// TypeEncoder is used by, and must be non-nil for, types which require type-resolution at Encode-Decode time.
	// That is, types which can reference new or unknown types.
	TypeEncoder Type

	// IncludeUnexported will include unexported struct fields in the encoded data.
	IncludeUnexported bool

	// used by pointer-types to resolve references;
	// i.e. multiple pointers to the same value, recursive references .. should all retain their reference structure.
	// It's no possible to resolve these things within the scope of a single type's Encodable, so here it is.
	// implementation handled entirely internally.
	r *referencer
}

// Encodable is an Encoder/Decoder for a specific type.
// They are typically not thread-safe, and passing a pointer to the wrong type results in undefined behaviour.
type Encodable interface {
	// Type returns the type that the Encodable encodes.
	// ptr in Encode() and Decode() must be a pointer to an object of this type.

	Type() reflect.Type

	// Encode writes the encoded form of the object at ptr to w.
	Encode(ptr unsafe.Pointer, w io.Writer) error

	// Decode reads the encoded form from r into the object at ptr.
	// Decodes will only read what Encode wrote.
	Decode(ptr unsafe.Pointer, r io.Reader) error

	// Size returns the maximum encoded size of the Encodable.
	// If Size returns <0, size is undefined.
	Size() int
}

// NewEncodable returns a new Encodable for encoding the type t.
// If the type is, or has a child type of interface, the TypeResolver tr will be used for encoding and decoding it.
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

	// Meta-Types
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

func (c *Config) copy() *Config {
	config := new(Config)
	*config = *c
	return config
}
