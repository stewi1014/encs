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
// TODO: Stress test recursive types and values.

import (
	"io"
	"reflect"
	"unsafe"
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
// Encodables return two kinds of error.
// encio.IOError for io and corrupted data errors, and encio.Error for encoding errors.
// See encs/encio/errors.go
//
// The pointers passed to Encode and Decode must be pointers to an allocated instance of the Encodable's type, accessible by Type().
// Pointer encodables do not follow different semantics, and so must be given a non-nil pointer to the pointer they're encoding.
// If the underlying pointer is nil, this is handled as it should be.
// See https://golang.org/pkg/unsafe/#Pointer; "Note that the pointer must point into an allocated object, so it may not be nil."
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

// New returns a new Encodable for encoding the type t.
// It uses DefaultSource as a Source.
func New(t reflect.Type, config Config) Encodable {
	return (&DefaultSource{}).NewEncodable(t, config)
}
