package encode

import (
	"fmt"
	"io"
	"reflect"
	"unsafe"

	"github.com/stewi1014/encs/encio"
	"github.com/stewi1014/encs/encodable"
	"github.com/stewi1014/encs/types"
)

// NewType returns a new reflect.Type encode.
func NewType(strict bool) *Type {
	e := &Type{
		unregistered: make(map[types.ID]reflect.Type),
		typeByID:     make(map[types.ID]reflect.Type),
		idByType:     make(map[reflect.Type]types.ID),
		buff:         make([]byte, 16),
		strict:       strict,
	}

	for _, t := range types.Registered {
		id := types.GetID(t)

		e.typeByID[id] = t
		e.idByType[t] = id
	}

	return e
}

// Type is an encodable for reflect.Type values.
type Type struct {
	// Strict type checking
	strict bool

	// unregistered contains types that should be known for speeds sake, but that have not been registered.
	// If an element in unregistered is used and neccecary for decoding, a warning should be shown, as the fact the type is
	// in unregistered is completely coincidental, and slight changes in usage or
	unregistered map[types.ID]reflect.Type
	typeByID     map[types.ID]reflect.Type

	// idByType contains both registered and unregistered types.
	idByType map[reflect.Type]types.ID
	buff     []byte
}

// Size implements Encodable.
func (e *Type) Size() int { return 16 }

// Type implements Encodable.
func (e *Type) Type() reflect.Type { return types.ReflectTypeType }

// Encode implements Encodable.
func (e *Type) Encode(ptr unsafe.Pointer, w io.Writer) error {
	ty := *(*reflect.Type)(ptr)
	id, ok := e.idByType[ty]
	if !ok {
		// This type hasn't been registered. Register it now.
		id = types.GetID(ty)

		e.idByType[ty] = id
	}

	id.Encode(e.buff)
	return encio.Write(e.buff, w)
}

// Decode implements Encodable.
// Apart from checking registered types, it also checks if the type currently in ptr is a match,
// and if so, does nothing.
func (e *Type) Decode(ptr unsafe.Pointer, r io.Reader) error {
	var id types.ID
	if err := encio.Read(e.buff, r); err != nil {
		return err
	}
	id.Decode(e.buff)

	ty, ok := e.typeByID[id]
	if ok {
		// Found the type
		*(*reflect.Type)(ptr) = ty
		return nil
	}

	ty, ok = e.unregistered[id]
	if ok {
		if ty == *(*reflect.Type)(ptr) {
			// Type is unregistered, but we're decoding into the same type as was sent.
			// Assume the caller know what it's doing. If it keeps only using us for validation instead of resolution
			// then we're good.
			return nil
		}

		// The received type is not registered, and the type in ptr is a different type.
		// We can resolve the type because we've seen it before, but we're relying on a fickle thing.
		// We must have been called with a ptr to this type previously.
		// Not only does this break the state-independent requirement of encs, but we're dependant on
		// the caller calling us in a certain order. I have half a mind to just throw an error instead here.

		fmt.Fprintln(encio.Warnings, encio.NewError(
			encio.ErrBadType,
			fmt.Sprintf(
				"%v is only known because it was previously passed to Decode; it isn't registered. this is unreliable and likely only worked out of pure luck. register it!",
				ty.String(),
			),
			0,
		))

		*(*reflect.Type)(ptr) = ty
		return nil
	}

	if *(*reflect.Type)(ptr) != nil {
		// No match, check if the existing type is a match.
		ty = *(*reflect.Type)(ptr)

		eid, ok := e.idByType[ty]
		if ok && eid == id {
			// fast path
			// Type is not registered, but the existing type has the same id as received.
			return nil
		}

		existing := types.GetID(ty)
		if existing == id {
			// slow path
			// Type is not registered, but the existing type has the same id as received.
			e.unregistered[id] = ty
			e.idByType[ty] = id
			return nil
		}

		// Neither existing type nor any registered types match strictly. Check loose types.
		var nameMatch reflect.Type
		if existing[0] == id[0] {
			// We have a loose match with the existing.
			if !e.strict {
				return nil
			}
			nameMatch = ty
		} else {
			for rty, rid := range e.idByType {
				if rid[0] == id[0] {
					if !e.strict {
						// We found a loose match with an existing type.
						*(*reflect.Type)(ptr) = rty
						return nil
					}
					nameMatch = rty
				}
			}
		}

		if nameMatch != nil && nameMatch.Name() != "" {
			return encio.NewError(encio.ErrBadType, fmt.Sprintf("unknown type ID %v. %v loosely matches, but LooseTyping isn't enabled.", id, nameMatch), 0)
		}
	}

	return encio.NewError(encio.ErrBadType, fmt.Sprintf("unknown type ID %v. Is it registered?", id), 0)
}

// NewValue returns a new reflect.Value encode.
func NewValue(src encodable.Source) *Value {
	return &Value{
		typeEnc: src.NewEncodable(types.ReflectTypeType, nil),
		src:     encodable.NewCachingSource(src),
		buff:    make([]byte, 1),
	}
}

// Value is an encodable for reflect.Value values.
type Value struct {
	typeEnc *encodable.Encodable
	src     *encodable.CachingSource
	buff    []byte
}

const (
	valueValid = 1 << iota
)

// Size implements Encodable.
func (e *Value) Size() int { return -1 << 31 }

// Type implements Encodable.
func (e *Value) Type() reflect.Type { return types.ReflectValueType }

// Encode implements Encodable.
func (e *Value) Encode(ptr unsafe.Pointer, w io.Writer) error {
	checkPtr(ptr)
	v := *(*reflect.Value)(ptr)

	if !v.IsValid() {
		e.buff[0] = 0
		return encio.Write(e.buff, w)
	}
	e.buff[0] = valueValid
	if err := encio.Write(e.buff, w); err != nil {
		return err
	}

	ty := v.Type()

	if !v.CanAddr() {
		// We can safely copy this value, as reflect informs us that this value is unaddressable.
		// reflect really better not be lying to us...
		// If it actually is addressable and reflect just doesn't want us to have the address then we may have a problem.
		// When encoding references I need one of two things. The address to compare with others, or a 100% guarantee that no other pointer *could* exist.
		// A user could always do some weird things (like me :\) with pointers and have a reference that shouldn't be possible in the go type system,
		// but they can write their own Encodable if that's the case.
		n := reflect.New(ty).Elem()
		n.Set(v)
		v = n
	}

	enc := e.src.NewEncodable(ty, nil)

	if err := (*e.typeEnc).Encode(unsafe.Pointer(&ty), w); err != nil {
		return err
	}

	return (*enc).Encode(unsafe.Pointer(v.UnsafeAddr()), w)
}

// Decode implements Encodable.
func (e *Value) Decode(ptr unsafe.Pointer, r io.Reader) error {
	checkPtr(ptr)

	if err := encio.Read(e.buff, r); err != nil {
		return err
	}

	if e.buff[0]&valueValid == 0 {
		*(*reflect.Value)(ptr) = reflect.Value{}
		return nil
	}

	var ty reflect.Type
	if err := (*e.typeEnc).Decode(unsafe.Pointer(&ty), r); err != nil {
		return err
	}

	enc := e.src.NewEncodable(ty, nil)

	v := (*reflect.Value)(ptr)
	if !v.CanAddr() || v.Type() != ty {
		n := reflect.New(ty).Elem()
		if v.IsValid() && ty == v.Type() {
			n.Set(*v)
		}
		*v = n
	}

	if err := (*enc).Decode(unsafe.Pointer(v.UnsafeAddr()), r); err != nil {
		// I no longer trust whatever is in the value
		*(*reflect.Value)(ptr) = reflect.Value{} // zero it
		return err
	}
	return nil
}
