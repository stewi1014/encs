// Package encodable provides low-level methods for seralising golang data structures.
// It aims to be fast, modular and comprehensive, valuing runtime speed over creation overhead.
//
// Encodable is the primary implementation, and provides Encode() and Decode() functions for a specific type.
package encodable

// I intend to keep a curated lsit of important notes to keep in mind while developing this part of encs here.
//
// Every instance of unsafe.Pointer that exists must always point towards a valid object
// unsafe.Poointer types are functional pointers with all the semantics that come with it;
// at any time, the garbage collector could come along, try to dereference the pointer and crash the program. yikes.
// If an invalid pointer is needed as an intermediary step, uintptr should be used.
//
// Recursive Types and recursive values are handled differently.
// Source is the solution to recursive types; it should stop a type trying to instantiate itself in its creation function,
// and return a placeholder.
// Recursive values are resolved by the Encodables themselves. E.g. pointer encodables should keep track of pointers they encode,
// and if they reach a cycle, encode a reference in the buffer that decoders can use to point the pointer to the previously decoded value.

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
	// LooseTyping will make encs ignore struct fields, interface methods and integer, float and complex sizes when resolving type equality.
	// TODO: Implement int, float and complex sizing flexibility.
	LooseTyping Config = 1 >> iota
)

const (
	// StructTag is the boolean struct tag that when applied to a struct, will force the fields inclusion or exclusion from encoding.
	// srvconv.ParseBool() is used for parsing the tag value; it accepts 1, t, T, TRUE, true, True, 0, f, F, FALSE, false, False.
	StructTag = "encs"
)

// Encodable is an Encoder and Decoder for a specific type.
//
// Encodables are not assumed to be thread safe.
// Encode() and Decode() often share static buffers for the sake of performance, but this comes at the cost of thread safety.
// Concurrent calls to either Encode() or Decode() will almost certainly result in complete failure.
// Use NewConcurrent or higher level functions if concurrency is needed.
//
// In a similar vein, Encodables are allowed to assume certain characteristics:
// 1. Calling Source.NewEncodable() will not result in infinite recursion, even if attempting to create the same Encodable.
// 1. Calling Encode() or Decode() on generated Encodable will not result in the calling Encodable instance being called again, even if it is the same Encodable.
// 1. Calling Encode() or Decode() on an unknown generated Encodable will not result in infite recursion **if pointers are equal**, even if it is the Same Encodable.
// 	If an encodable copies the value it passes to its elemtn Encodables, then pointers cannot be compared, and
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

// NewFunc is the function signature of functions that create new Encodables.
type NewFunc func(reflect.Type, Config, Source) Encodable

// New returns a new Encodable for encoding the type ty.
func New(ty reflect.Type, config Config, src Source) Encodable {
	ptrt := reflect.PtrTo(ty)
	kind := ty.Kind()
	switch {
	// Implementers
	case ptrt.Implements(binaryMarshalerType) && ptrt.Implements(binaryUnmarshalerType):
		return NewBinaryMarshaler(ty)

	// Specific types
	case ty == reflectTypeType:
		return NewType(config)
	case ty == reflectValueType:
		return NewValue(config, src)

	// Compound-Types
	case kind == reflect.Ptr:
		return NewPointer(ty, config, src)
	case kind == reflect.Interface:
		return NewInterface(ty, config, src)
	case kind == reflect.Struct:
		return NewStruct(ty, config, src)
	case kind == reflect.Array:
		return NewArray(ty, config, src)
	case kind == reflect.Slice:
		return NewSlice(ty, config, src)
	case kind == reflect.Map:
		return NewMap(ty, config, src)

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
	default:
		panic(encio.NewError(encio.ErrBadType, fmt.Sprintf("cannot create encodable for type %v", ty), 0))
	}
}
