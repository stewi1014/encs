package encodable

import (
	"fmt"
	"io"
	"reflect"
	"unsafe"

	"github.com/stewi1014/encs/encio"
)

// EncID is comparable struct representing an encodable of a particular type and configuration.
// It can be used as keys in maps of Encodables to confirm equality across Encodables.
type EncID struct {
	reflect.Type
	Config
}

// Source is a generator of Encodables. It is typically used as a method for avoiding infinite recursion on recursive types,
// but can also serve to modify the default behaviour of encodable generation. Compound type Encodables take Source as an argument upon creation,
// and use it for the generation of their element types either during creation, or they keep it and use it for encode-time generation.
type Source interface {
	// NewEncodable returns a new Encodable.
	// Source should take care to avoid infinite recursion, taking note of when it is called to create an Encodable from inside the same Encodable's creation function.
	// Recursive provides a placeholder Encodable for types that would otherwise recurse, deferring the creation of the needed Encodable to when Encoding is actually taking place.
	NewEncodable(reflect.Type, Config) Encodable
}

// NewCachingSource returns a new CachingSource, using source for cache misses.
// Users of CachingSource must not pass it to element Encodables, else a situation may arise where
// a recursive type Encodable that's generated at encode-time is given itself from the cache, and it makes
// nested calls to itself. Use CachingSource.Source to pass to element Encodables.
func NewCachingSource(source Source) *CachingSource {
	return &CachingSource{
		cache:  make(map[EncID]Encodable),
		Source: source,
	}
}

// CachingSource provides a cache of Encodables.
type CachingSource struct {
	cache map[EncID]Encodable
	Source
}

// NewEncodable implements Source.
func (src *CachingSource) NewEncodable(ty reflect.Type, config Config) (enc Encodable) {
	enc, ok := src.cache[EncID{Type: ty, Config: config}]
	if ok {
		return enc
	}

	enc = src.Source.NewEncodable(ty, config)
	src.cache[EncID{Type: ty, Config: config}] = enc
	return enc
}

// DefaultSource safely creates Encodables for recursive types.
// It provides no further functionality over the resolution of recursive types,
// and creation of Encodables in this library.
// It can be initialised with DefaultSource{}.
type DefaultSource map[EncID]*Recursive

// NewEncodable implements Source.
// When NewEncodable is called to make an Encodable that is currently being created,
// it returns a Recursive Encodable instead.
func (src *DefaultSource) NewEncodable(ty reflect.Type, config Config) (enc Encodable) {
	if *src == nil {
		m := make(DefaultSource)
		src = &m
	}

	id := EncID{
		Type:   ty,
		Config: config,
	}

	if r, ok := (*src)[id]; ok {
		// We've been asked to create an encodable that's already in the process of being created; a recursive type.
		if r != nil {
			// Already have one
			return r
		}

		r := NewRecursive(func() Encodable {
			return src.makeEncodable(ty, config)
		})
		(*src)[id] = r
		return r
	}
	// Need to make a new encodable

	(*src)[id] = nil // Let ourselves know we're currently building an encodable of this type.

	enc = src.makeEncodable(ty, config)

	// We're done making the encodable.
	if r, ok := (*src)[id]; ok && r != nil {
		enc = r
	}

	return
}

// makeEncodable creates an encodable with no checks.
// It is used either after recursive checks have taken place, or in the generation function passed to Recursive.
func (src *DefaultSource) makeEncodable(ty reflect.Type, config Config) Encodable {
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

// RecursiveMaxCache is the maximum number of Encodables that Recursive will cache.
const RecursiveMaxCache = 256

// NewRecursive returns a new Recursive Encodable.
func NewRecursive(newFunc func() Encodable) *Recursive {
	return &Recursive{
		new:  newFunc,
		encs: make([]Encodable, 0, 1),
	}
}

// Recursive is an encodable that only creates its wrapped Encodable when called.
// It caches instances to avoid creation on every call.
type Recursive struct {
	new  func() Encodable
	encs []Encodable
}

// Size implements Encodable.
// Recursive types have a theoretical infinite size.
func (e *Recursive) Size() int {
	return -1 << 31
}

// Type implements Encodable.
func (e *Recursive) Type() reflect.Type {
	enc := e.Pop()
	defer e.Push(enc)
	return enc.Type()
}

// Encode implements Encodable.
func (e *Recursive) Encode(ptr unsafe.Pointer, w io.Writer) error {
	enc := e.Pop()
	defer e.Push(enc)
	return enc.Encode(ptr, w)
}

// Decode implements Encodable.
func (e *Recursive) Decode(ptr unsafe.Pointer, r io.Reader) error {
	enc := e.Pop()
	defer e.Push(enc)
	return enc.Decode(ptr, r)
}

// Pop takes an encodable off the top of the stack or generates one,
// passing ownership to the caller.
func (e *Recursive) Pop() (enc Encodable) {
	l := len(e.encs)
	if l > 0 {
		enc = e.encs[l-1]
		e.encs = e.encs[:l-1]
		return
	}

	return e.new()
}

// Push returns an encodable to the stack,
// taking ownership of it. Subsequent calls cannot be made to the Encodable.
func (e *Recursive) Push(enc Encodable) {
	l := len(e.encs)
	if l < cap(e.encs) {
		e.encs = e.encs[:l+1]
		e.encs[l] = enc
		return
	}

	nb := make([]Encodable, l+1, cap(e.encs)*2)
	copy(nb, e.encs)
	nb[l] = enc
	e.encs = nb
}
