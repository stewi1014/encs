package encodable

import (
	"errors"
	"fmt"
	"io"
	"reflect"
	"unsafe"

	"github.com/stewi1014/encs/encio"
)

// This file contains the logic for handling recursive types and values.

// NewRecursiveSource returns a new RecursiveSource.
//
// The provided Source can be 'dumb'; i.e. A big switch statement to create an Encodable for a type.
// It must respect the implementation details of Source; if it doesn't pass the source passed when creating an encodable,
// RecursiveSource cannot resolve recursive types.
func NewRecursiveSource(source Source) *RecursiveSource {
	return &RecursiveSource{
		source: source,
		seen:   make(map[reflect.Type]*genState),
		ptrs:   NewPointers(),
	}
}

// RecursiveSource safely creates Encodables for recursive types.
type RecursiveSource struct {
	source Source
	ptrs   Pointers
	seen   map[reflect.Type]*genState
	hasAny bool
	depth  int
}

// genState is the state of generation of an encodable.
type genState struct {
	enc    *Encodable // changes to this field must be done through de-referencing.
	source Source     // Let's not assume Sources can't change during generation; make sure we use the same one.
	depth  int
	state  byte
}

const (
	stateGenerating = iota
	stateGenerated
	stateRecursed
)

// NewEncodable implements Source.
// It manages how recursive Encodables are generated, and substitutes with Recursive where needed.
// The returned Encodable is usable upon return, but can be retroactively swapped. As such the returned Encodable must not be dereferenced.
// It wraps Encodables as little as possible while still keeping the ability to handle recursive values and types,
// and the correct re-creation of recursive values on decode.
func (src *RecursiveSource) NewEncodable(ty reflect.Type, source Source) *Encodable {
	if source == nil {
		source = src
	}

	src.depth++
	defer func() { src.depth-- }()

	id := ty

	if ty.Kind() == reflect.Interface || ty == reflectValueType {
		// These types can contain anything.
		src.hasAny = true
		src.setRecursed(0)
	}

	if src.hasAny {
		return src.makeRecursive(id, source)
	}

	if gen, ok := src.seen[id]; ok {
		if gen.state == stateRecursed || gen.state == stateGenerated {
			// Here's one I made earlier.
			return gen.enc
		}

		if gen.state == stateGenerating {
			// Recursive type!
			src.setRecursed(gen.depth)
			return src.makeRecursive(id, gen.source)
		}

		// Radiation from the sun flipping bits???
		panic("invalid generation state")
	}

	// Need to make one

	// Let ourselves know what we're doing
	src.seen[id] = &genState{
		enc:    new(Encodable),
		source: source,
		depth:  src.depth,
		state:  stateGenerating,
	}

	enc := src.source.NewEncodable(ty, source)

	// Time to see what happened.

	gen, ok := src.seen[id]
	if !ok {
		// This one is definitely impossible. We never call delete or re-allocate the map.
		panic("encodable removed from map")
	}

	if gen.state == stateRecursed {
		// The Encodable recursed.
		// Not needed, but it's nice to give Recursive the Encodable we did make.
		// The first encode or decode should be 500ns faster for what it's worth.
		(*gen.enc).(*Recursive).Push(enc)
		return gen.enc
	}

	if gen.state == stateGenerating {
		// No recurse
		*gen.enc = *enc
		gen.state = stateGenerated
		return gen.enc
	}

	// Only other option is stateGenerated.
	// Well excuse me. I decide when this encodable has finished being made, because I'm making it.
	panic("invalid generation state")
}

// setRecursed makes all the currently generating encodables with a depth greater than depth recursive.
// It's used when a type is found to be recursive; all the element types of it must be wrapped with recursive,
// as we can't allow nested calls to the same Encodable instance.
func (src *RecursiveSource) setRecursed(depth int) {
	for id, gen := range src.seen {
		if gen.depth <= depth || gen.state != stateGenerating {
			continue
		}

		src.makeRecursive(id, gen.source)
	}
}

// makeRecursive either wraps an existing Encodable with Recursive, or creates a new Recursive Encodable.
// It updates the map.
func (src *RecursiveSource) makeRecursive(id reflect.Type, source Source) *Encodable {
	if gen, ok := src.seen[id]; ok {
		if gen.state == stateRecursed {
			return gen.enc
		}

		recursive := NewRecursive(id, &src.ptrs, func() *Encodable {
			return src.source.NewEncodable(id, gen.source)
		})

		*gen.enc = recursive
		gen.state = stateRecursed

		return gen.enc
	}

	enc := Encodable(NewRecursive(id, &src.ptrs, func() *Encodable {
		return src.source.NewEncodable(id, source)
	}))

	src.seen[id] = &genState{
		enc:    &enc,
		source: source,
		depth:  src.depth,
		state:  stateRecursed,
	}

	return &enc
}

// NewPointers returns a new Pointers.
func NewPointers() Pointers {
	return Pointers{
		Int32: encio.NewInt32(),
	}
}

// Pointers provides methods for keeping track of pointers and their types, it is helpful for detecting and resolving pointer cycles.
// It takes a reflect.Type in methods, sometimes uneccecaraly, but allows it to asset the correct type before return.
// returning an error if types do not match. This is not so neccecary for Encoding, but is very important to validate when Decoding.
//
// Encoders can call Has() before encoding to check if the pointer has already been encoded,
// and decoders can call Get() to retrieve a previously decoded object. Both must call Add() if they do the real Encode/Decode.
//
// Encodables should call First(), and if they are the first caller, defer a call to Reset().
// This clears the pointer slice between encodes/decodes.
type Pointers struct {
	pointers []object
	in       bool
	encio.Int32
}

// object contains information about a value.
// The value in ptr must always be a valid instance of the given type.
type object struct {
	ptr unsafe.Pointer
	reflect.Type
}

// Get returns the object at the given index.
func (p *Pointers) Get(index int32, ty reflect.Type) (unsafe.Pointer, error) {
	if index < 0 || index >= int32(len(p.pointers)) {
		return nil, fmt.Errorf(
			"index out of bounds, object %v is referenced but we only have %v objects",
			index,
			len(p.pointers),
		)
	}

	if p.pointers[index].Type != ty {
		return nil, fmt.Errorf(
			"object %v referenced, but it has type %v. wanted %v",
			index,
			p.pointers[index].Type.String(),
			ty.String(),
		)
	}

	return p.pointers[index].ptr, nil
}

// Add adds the pointer to the list of known pointers.
func (p *Pointers) Add(ptr unsafe.Pointer, ty reflect.Type) {
	l := len(p.pointers)
	if l < cap(p.pointers) {
		p.pointers = p.pointers[:l+1]
		p.pointers[l] = object{ptr: ptr, Type: ty}
		return
	}

	var nb []object
	if cap(p.pointers) == 0 {
		nb = make([]object, l+1, 2)
	} else {
		nb = make([]object, l+1, cap(p.pointers)*2)
	}
	copy(nb, p.pointers)
	nb[l] = object{ptr: ptr, Type: ty}
	p.pointers = nb
}

// Has returns true if the given pointer has already been added.
func (p *Pointers) Has(ptr unsafe.Pointer, ty reflect.Type) (index int32, ok bool) {
	for i := int32(0); i < int32(len(p.pointers)); i++ {
		if p.pointers[i].ptr == ptr && p.pointers[i].Type == ty {
			return i, true
		}
	}
	return 0, false
}

// HasReference returns true and the index if the given reference type's pointer matches an existing value.
// It panics if v's kind is not Chan, Func, Map, Ptr, Slice, or UnsafePointer.
func (p *Pointers) HasReference(v reflect.Value) (int32, bool) {
	for i := int32(0); i < int32(len(p.pointers)); i++ {
		if p.pointers[i].Type == v.Type() {
			if reflect.NewAt(v.Type(), p.pointers[i].ptr).Elem().Pointer() == v.Pointer() {
				return i, true
			}
		}
	}
	return 0, false
}

// First returns true if this is the first call to First() since the last call to Reset().
// If it is, the encodable should call Reset() when it exits.
func (p *Pointers) First() bool {
	if p.in {
		return false
	}

	p.in = true
	return true
}

// Reset removes all objects. Pointer can be reused for Encoding/Decoding after a call to Reset.
func (p *Pointers) Reset() {
	p.pointers = p.pointers[:0]
	p.in = false
}

// RecursiveMaxCache is the maximum number of Encodables that Recursive will cache.
const RecursiveMaxCache = 256

// NewRecursive returns a new Recursive Encodable.
// ptrs should be unique between element encodables e.g. a struct Encodable should share ptrs with Encodables for its fields.
// This sharing should be handled by Source.
//
// Typically, it is never desirable to manually create an instance of Recursive. It is best used inside implementations of Source
// as a solution to recursive types and values.
func NewRecursive(ty reflect.Type, ptrs *Pointers, newFunc func() *Encodable) *Recursive {
	r := &Recursive{
		ptrs: ptrs,
		ty:   ty,
		new:  newFunc,
		encs: make([]*Encodable, 0, 1),
	}

	if ty.Kind() == reflect.Map {
		r.kind = rReference
	}

	return r
}

// recursive kinds.
const (
	rReference = 1 << iota
)

// encoded meanings.
const (
	ptrEncoded = -1
)

// Recursive is an encodable that resolves recursive values and types.
// It aims to reproduce reference cycles with 100% accuracy; e.g. if two map values reference the same underlying map,
// they are decoded to reference the same map.
// Recursive is best used inside implementations of Source as a solution to recursive types and values.
//
// It resolves Encodable creation for recursive types by deferring encodable creation until Encode/Decode.
//
// It resolves the nested calling of Encodables by keeping a buffer of instances of the underying Encodable. Calling an instance of Recursive twice will result in
// two distinct instances of the underlying being called.
//
// Recursive values are resolved through a shared instance of Pointers between all Recursive Encodables for a type.
// During Encode, Recursive will check if the pointer has already been encoded, taking care to check reference types such as maps.
// If not, it then records the pointer, and encodes using the underlying Encodable returned from newFunc.
//

type Recursive struct {
	new  func() *Encodable
	ty   reflect.Type
	ptrs *Pointers
	kind byte
	encs []*Encodable
}

// Size implements Encodable.
// Recursive types have a theoretical infinite maximum size.
func (e *Recursive) Size() int { return -1 << 31 }

// Type implements Encodable.
func (e *Recursive) Type() reflect.Type { return e.ty }

// Encode implements Encodable.
func (e *Recursive) Encode(ptr unsafe.Pointer, w io.Writer) error {
	if e.ptrs.First() {
		// We're the first call. Clear pointers when we exit.
		defer e.ptrs.Reset()
	}

	index, has := e.ptrs.Has(ptr, e.ty)

	if has {
		// We've already encoded this pointer.
		return e.ptrs.Encode(w, index)
	}

	// Check reference types.
	if val := reflect.NewAt(e.ty, ptr).Elem(); e.kind == rReference && !val.IsNil() {
		if index, has := e.ptrs.HasReference(val); has {
			return e.ptrs.Encode(w, index)
		}
	}

	e.ptrs.Add(ptr, e.ty)
	if err := e.ptrs.Encode(w, ptrEncoded); err != nil {
		return err
	}

	enc := e.Pop()
	defer e.Push(enc)
	return (*enc).Encode(ptr, w)
}

// Decode implements Encodable.
func (e *Recursive) Decode(ptr unsafe.Pointer, r io.Reader) error {
	if e.ptrs.First() {
		// We're the first call. Clear pointers when we exit.
		defer e.ptrs.Reset()
	}

	index, err := e.ptrs.Decode(r)
	if err != nil {
		return err
	}

	if index >= 0 {
		previous, err := e.ptrs.Get(index, e.ty)
		if err != nil {
			return err
		}

		reflect.NewAt(e.ty, ptr).Elem().Set(reflect.NewAt(e.ty, previous).Elem())
		return nil
	}

	e.ptrs.Add(ptr, e.ty)

	enc := e.Pop()
	defer e.Push(enc)
	return (*enc).Decode(ptr, r)
}

// Pop takes an encodable off the top of the stack or generates one,
// passing ownership to the caller.
func (e *Recursive) Pop() (enc *Encodable) {
	l := len(e.encs)
	if l > 0 {
		enc = e.encs[l-1]
		e.encs = e.encs[:l-1]
		return
	}

	enc = e.new()
	if _, ok := (*enc).(*Recursive); ok {
		panic(encio.NewError(errors.New("newFunc must not return a *Recursive encodable"), "trying to create the underying Encodable in Recursive just returned another Recursive instance", 0))
	}
	return
}

// Push returns an encodable to the stack,
// taking ownership of it. Subsequent calls cannot be made to the Encodable.
func (e *Recursive) Push(enc *Encodable) {
	l := len(e.encs)
	if l < cap(e.encs) {
		e.encs = e.encs[:l+1]
		e.encs[l] = enc
		return
	}

	nb := make([]*Encodable, l+1, cap(e.encs)*2)
	copy(nb, e.encs)
	nb[l] = enc
	e.encs = nb
}
