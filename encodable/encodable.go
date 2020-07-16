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
	// LooseTyping will make encs ignore struct fields, interface methods and integer, float and complex sizes when resolving type equality.
	// TODO: Implement int, float and complex sizing flexibility.
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
//
// In order to provide these assumptions, Encodables must
// 1. Never copy the value they pass to element Encodables, with the exception of Map types, and encodables of the reflect.Value type.
// 	Things seem to work ok most of the time when doing this, but it's not a good idea.
// 1. Only create new Encodables using the provided Source.
// 1. Pass the same Source they were given to element Encodables.
// 1. Never pass Encodables to other Encodables; each Encodable must make its own element Encodables.
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

// DefaultSource is a simple Source for Encodables. It performs no recursion checks.
// Use RecursiveSource unless it is guaranteed there will not be recursive types or values encoded.
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

	// Integer types
	case kind == reflect.Uint8:
		*enc = NewUint8()
	case kind == reflect.Uint16:
		*enc = NewUint16()
	case kind == reflect.Uint32:
		*enc = NewUint32()
	case kind == reflect.Uint64:
		*enc = NewUint64()
	case kind == reflect.Uint:
		*enc = NewUint()
	case kind == reflect.Int8:
		*enc = NewInt8()
	case kind == reflect.Int16:
		*enc = NewInt16()
	case kind == reflect.Int32:
		*enc = NewInt32()
	case kind == reflect.Int64:
		*enc = NewInt64()
	case kind == reflect.Int:
		*enc = NewInt()
	case kind == reflect.Uintptr:
		*enc = NewUintptr()

	// Float types
	case kind == reflect.Float32:
		*enc = NewFloat32()
	case kind == reflect.Float64:
		*enc = NewFloat64()
	case kind == reflect.Complex64:
		*enc = NewComplex64()
	case kind == reflect.Complex128:
		*enc = NewComplex128()

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
