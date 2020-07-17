// Package encodable provides low-level methods for seralising golang data structures.
// It aims to be fast, modular and comprehensive, valuing runtime speed over creation overhead.
//
// Encodable is the primary implementation, and provides Encode() and Decode() functions for a specific type.
package encodable

// I intend to keep a curated lsit of important notes to keep in mind while developing this part of encs here.
//
// https://golang.org/pkg/unsafe/#Pointer; "Note that the pointer must point into an allocated object, so it may not be nil".
// Every instance of unsafe.Pointer that exists must always point towards a valid object because apparently it can cause the garbage collector to panic.
// Except, it doesn't. nil values of unsafe.Pointer are used in the standard libraries.
//
// Recursion detection relises on the same pointer being attempted to Encode from. Encodables may allocate and pass whatever pointers
// they want to element Decoders when decoding, but they may not do so during encoding. If and Encodable allocates a new object and passes it to its element encodables,
// then it must implement its own recursion checks; the Recursive Encodable will not work.

import (
	"fmt"
	"io"
	"reflect"
	"unsafe"

	"github.com/stewi1014/encs/encio"
)

// Config controls the configuration of encodables.
type Config int

const (
	// LooseTyping will make encs ignore struct fields in named structs and interface methods,
	// number types (int*, uint* and float*) are treated as equal,
	// and complex types (complex64, complex128) are treated as equal.
	// For implementation details see StructLoose, Varint, and VarComplex encodables, and for exact details on how type resolving is changed, see Type.
	LooseTyping Config = 1 >> iota

	// LogTypes will make Type Encodables log types and their generated ids when encoding.
	// It is helpful for debugging a type that cannot be resolved on Decode.
	LogTypes
)

const (
	// StructTag is the boolean struct tag that when applied to a struct, will force the field's inclusion or exclusion from encoding.
	// srvconv.ParseBool() is used for parsing the tag value; it accepts 1, t, T, TRUE, true, True, 0, f, F, FALSE, false, False.
	StructTag = "encs"
)

// Encodable is an Encoder and Decoder for a specific type.
//
// Encodables are not assumed to be thread safe.
// Use NewConcurrent or higher level functions if concurrency is needed.
//
// In a similar vein, Encodables are allowed to assume certain characteristics:
// 1. Encode() and Decode() will not be called concurrently.
// 1. Calling NewEncodable() on the provided Source will not result in infinite recursion, even if attempting to create the same Encodable.
// 1. Calling Encode() or Decode() on an Encodable from the provided Source will not result in the calling Encodable instance being called again, even if it is the same Encodable.
// 1. Calling Encode() on an Encodable from the provided Source will not result in infinite recursion unless the passed pointer is to a copied object.
// 1. Calling Decode() on an Encodable from the provided Source will not result in infinite recursion.
// 1. Reading from the io.Reader provided to Decode will return the same data, in the same order as was written to the io.Writer in Encode.
//
// In order to provide these assumptions, Encodables must
// 1. Never copy the value they pass to element Encodables, with the exception of Map types, and encodables of the reflect.Value type.
// 	Things seem to work ok most of the time when doing this, but it's not a good idea.
// 1. Only create element Encodables using the provided source. An element encodable is any encodable that has a different type, or is called with a different pointer address.
// 	i.e. An int encodable could create its own int64 Encodable to delegate to, but a struct Encodable may not create its own Encodables for fields, even the first field with the same pointer address.
// 1. Always pass the same Source they were given to element Encodables.
// 1. Never pass Encodables to other Encodables; each Encodable must make its own element Encodables.
// 1. Always read the same amount of data as was written, even if the data is useless. Don't leave garbage in the buffer for the next encodable.
//
// Encodables return two kinds of error.
// encio.IOError for io and corrupted data errors, and encio.Error for encoding errors.
// See encs/encio/errors.go
//
// The pointers passed to Encode and Decode must be pointers to an allocated instance of the Encodable's type, accessible by Type().
// Pointer encodables do not follow different semantics, and so must be given a non-nil pointer to the pointer they're encoding.
// If the underlying pointer is nil, this is handled as it should be.
// See https://golang.org/pkg/unsafe/#Pointer; "Note that the pointer must point into an allocated object, so it may not be nil".
type Encodable interface {
	// Type returns the type that the Encodable encodes.
	// It is thread safe.
	Type() reflect.Type

	// Size returns the maximum encoded size of the Encodable.
	// If Size returns <0, size is undefined.
	// It is thread safe.
	Size() int

	// Encode encodes the object at ptr to w.
	// If Size() returns >1, Encode will write at most Size() bytes.
	// It panics if ptr is nil.
	Encode(ptr unsafe.Pointer, w io.Writer) error

	// Decode decodes from r into the object at ptr.
	// Decode will only read what Encode wrote; no extra data is read.
	// It panics if ptr is nil.
	Decode(ptr unsafe.Pointer, r io.Reader) error
}

// NewDefaultSource returns a new DefaultSoure.
func NewDefaultSource() DefaultSource { return DefaultSource{} }

// DefaultSource is a simple Source for Encodables. It performs no pointer logic.
// Use RecursiveSource unless it is guaranteed there will not be recursive types or values encoded,
// and the pointer reference structure doesn't matter. i.e. If a struct Encodable is created with an int and *int field
// where the *int field points to the int field, the decoded *int field will not point to the struct's own field.
// It is also slower for large types.
// DefaultSource{} is an appropriate way to instantiate it.
type DefaultSource struct{}

// NewEncodable implements Source.
func (s DefaultSource) NewEncodable(ty reflect.Type, config Config, src Source) (enc *Encodable) {
	if src == nil {
		src = s
	}

	enc = new(Encodable)

	ptrt := reflect.PtrTo(ty)
	kind := ty.Kind()
	switch {
	// Implementers
	case ptrt.Implements(binaryMarshalerType) && ptrt.Implements(binaryUnmarshalerType):
		*enc = NewBinaryMarshaler(ty)

	// Specific types
	case ty == reflectTypeType:
		*enc = NewType(config)
	case ty == reflectValueType:
		*enc = NewValue(config, src)

	// Compound-Types
	case kind == reflect.Ptr:
		*enc = NewPointer(ty, config, src)
	case kind == reflect.Interface:
		*enc = NewInterface(ty, config, src)
	case kind == reflect.Struct:
		*enc = NewStruct(ty, config, src)
	case kind == reflect.Array:
		*enc = NewArray(ty, config, src)
	case kind == reflect.Slice:
		*enc = NewSlice(ty, config, src)
	case kind == reflect.Map:
		*enc = NewMap(ty, config, src)

	// Number types
	case kind == reflect.Uint8,
		kind == reflect.Uint16,
		kind == reflect.Uint32,
		kind == reflect.Uint64,
		kind == reflect.Uint,
		kind == reflect.Int8,
		kind == reflect.Int16,
		kind == reflect.Int32,
		kind == reflect.Int64,
		kind == reflect.Int,
		kind == reflect.Uintptr,
		kind == reflect.Float32,
		kind == reflect.Float64:
		*enc = NewNumber(ty, config)

	case kind == reflect.Complex64,
		kind == reflect.Complex128:
		*enc = NewComplex(ty, config)

	// Misc types
	case kind == reflect.Bool:
		*enc = NewBool()
	case kind == reflect.String:
		*enc = NewString()
	default:
		panic(encio.NewError(encio.ErrBadType, fmt.Sprintf("cannot create encodable for type %v", ty), 0))
	}

	return
}

// NewNumber generates an encodable for the given number type.
// It supports the LooseTyping gflag, and if set, returns Varint
// an encodable that can encode from/decode to any int* uint* and float* type.
// Otherwise, it returns the appropriate type specific Encodable.
func NewNumber(ty reflect.Type, config Config) Encodable {
	kind := ty.Kind()
	switch {
	case config&LooseTyping != 0:
		return NewVarint(ty)
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
	case kind == reflect.Float32:
		return NewFloat32()
	case kind == reflect.Float64:
		return NewFloat64()
	default:
		panic(encio.NewError(encio.ErrBadType, fmt.Sprintf("%v is not a number type. must be int* uint*, or float*", ty.String()), 0))
	}
}
