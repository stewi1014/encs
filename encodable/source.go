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

// Source is a generator of Encodables. Compound type Encodables take Source as an argument upon creation,
// and use it for the generation of their element types either during creation or for encode-time generation.
// The Source is responsible for resolving recursive types.
type Source interface {
	// NewEncodable returns a new Encodable.
	// Source should take care to avoid infinite recursion, taking note of when it is called to create an Encodable from inside the same Encodable's creation function.
	// Recursive provides a placeholder Encodable for types that would otherwise recurse, deferring the creation of the needed Encodable to when Encoding is actually taking place.
	NewEncodable(reflect.Type, Config) Encodable
}

// NewCachingSource returns a new CachingSource, using source for cache misses.
// Users of CachingSource must not pass it to element Encodables why may try to create themselves,
// else a situation may arise where a recursive type Encodable is given itself from the cache, and it makes
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

// NewRecursiveSource returns a new RecursiveSource, using newFunc as a source for encodables.
// Typically, encodable.New is passed as newFunc, but any encodable creating function can be passed to
// customise encodable generation.
func NewRecursiveSource(newFunc NewFunc) *RecursiveSource {
	return &RecursiveSource{
		new:  newFunc,
		seen: make(map[EncID]*Recursive),
	}
}

// RecursiveSource safely creates Encodables for recursive types.
// It returns Recursive when resolving recursive types.
type RecursiveSource struct {
	new  NewFunc
	seen map[EncID]*Recursive
}

// NewEncodable implements Source.
// When NewEncodable is called to make an Encodable that is currently being created,
// it returns a Recursive Encodable instead.
func (src *RecursiveSource) NewEncodable(ty reflect.Type, config Config) (enc Encodable) {
	id := EncID{
		Type:   ty,
		Config: config,
	}

	if r, ok := src.seen[id]; ok {
		// We've been asked to create an encodable that's already in the process of being created; a recursive type.
		if r != nil {
			// Already have one
			return r
		}

		r := NewRecursive(ty, func() Encodable {
			return New(ty, config, src)
		})
		src.seen[id] = r
		return r
	}
	// Need to make a new encodable

	src.seen[id] = nil // Let ourselves know we're currently building an encodable of this type.

	enc = New(ty, config, src)

	// We're done making the encodable.
	r, ok := src.seen[id]
	if r != nil {
		// This was a recursive-type Encodable. Replace our returned Encodable with Recursive so it can maintain all cycles.
		r.Push(enc)
		enc = r
	} else if ok {
		// This was not a recursive type, delete it.
		delete(src.seen, id)
	}

	return
}

// RecursiveMaxCache is the maximum number of Encodables that Recursive will cache.
const RecursiveMaxCache = 256

// NewRecursive returns a new Recursive Encodable.
func NewRecursive(ty reflect.Type, newFunc func() Encodable) *Recursive {
	r := &Recursive{
		ty:   ty,
		new:  newFunc,
		encs: make([]Encodable, 0, 1),
	}

	switch ty.Kind() {
	case reflect.Chan,
		reflect.Func,
		reflect.Map,
		reflect.Ptr,
		reflect.Slice,
		reflect.UnsafePointer:
		r.state |= recursiveReferenceType
	}

	return r
}

const (
	recursiveMaster = 1 << iota
	recursiveReferenceType
)

// Recursive is an encodable that resolved recursive values and types.
//
// Instances of Encodable are allowed to assume that calls to unknown element Encodables will not result in
// a nested call to themselves
// It avoids the creating of the wrapped Encodable
type Recursive struct {
	new      func() Encodable
	ty       reflect.Type
	intEnc   encio.Int
	state    byte
	pointers []unsafe.Pointer
	encs     []Encodable
}

// Size implements Encodable.
// Recursive types have a theoretical infinite size.
func (e *Recursive) Size() int {
	return -1 << 31
}

// Type implements Encodable.
func (e *Recursive) Type() reflect.Type { return e.ty }

// Encode implements Encodable.
func (e *Recursive) Encode(ptr unsafe.Pointer, w io.Writer) error {
	if e.master() {
		// We're the first call. Clear pointers when we exit.
		defer e.reset()
	}

	index, has := e.has(ptr)
	if has {
		// We've already encoded this pointer.
		return e.intEnc.EncodeInt32(w, index)
	}

	if err := e.intEnc.EncodeInt32(w, -1); err != nil {
		return err
	}

	e.add(ptr)

	enc := e.Pop()
	defer e.Push(enc)
	return enc.Encode(ptr, w)
}

// Decode implements Encodable.
func (e *Recursive) Decode(ptr unsafe.Pointer, r io.Reader) error {
	if e.master() {
		// We're the first call. Clear pointers when we exit.
		defer e.reset()
	}

	index, err := e.intEnc.DecodeInt32(r)
	if err != nil {
		return err
	}

	if index >= 0 {
		// Encoded by reference
		if index >= int32(len(e.pointers)) {
			return encio.NewError(
				encio.ErrMalformed,
				fmt.Sprintf("pointer reference %v is out of bounds; only have %v references", index, len(e.pointers)),
				0,
			)
		}

		reflect.NewAt(e.ty, ptr).Elem().Set(reflect.NewAt(e.ty, e.pointers[index]).Elem())
		return nil
	}

	e.add(ptr)

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

func (e *Recursive) master() bool {
	if e.state&recursiveMaster > 0 {
		return false
	}

	e.state |= recursiveMaster
	return true
}

// Reset should be called at the end of every encode session, else old pointer data could be re-used across encodes.
func (e *Recursive) reset() {
	e.pointers = e.pointers[:0]
	e.state &^= recursiveMaster
}

func (e *Recursive) add(ptr unsafe.Pointer) {
	l := len(e.pointers)
	if l < cap(e.pointers) {
		e.pointers = e.pointers[:l+1]
		e.pointers[l] = ptr
		return
	}

	var nb []unsafe.Pointer
	if cap(e.pointers) == 0 {
		nb = make([]unsafe.Pointer, l+1, 2)
	} else {
		nb = make([]unsafe.Pointer, l+1, cap(e.pointers)*2)
	}
	copy(nb, e.pointers)
	nb[l] = ptr
	e.pointers = nb
}

func (e *Recursive) has(ptr unsafe.Pointer) (int32, bool) {
	for i := int32(0); i < int32(len(e.pointers)); i++ {
		if e.pointers[i] == ptr {
			return i, true
		}
		if e.state&recursiveReferenceType > 0 {
			existing := reflect.NewAt(e.ty, e.pointers[i]).Elem()
			val := reflect.NewAt(e.ty, ptr).Elem()
			if existing.Pointer() == val.Pointer() {
				return i, true
			}
		}
	}

	return 0, false
}
