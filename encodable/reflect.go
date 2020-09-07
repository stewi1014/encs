package encodable

import (
	"errors"
	"fmt"
	"io"
	"reflect"
	"time"
	"unsafe"

	"github.com/stewi1014/encs/encio"
)

var (
	reflectTypeType  = reflect.TypeOf(new(reflect.Type)).Elem()
	reflectValueType = reflect.TypeOf(new(reflect.Value)).Elem()
)

func init() {
	err := Register(
		nil,
		reflect.TypeOf(int(0)),
		reflect.TypeOf(int8(0)),
		reflect.TypeOf(int16(0)),
		reflect.TypeOf(int32(0)),
		reflect.TypeOf(int64(0)),
		reflect.TypeOf(uint(0)),
		reflect.TypeOf(uint8(0)),
		reflect.TypeOf(uint16(0)),
		reflect.TypeOf(uint32(0)),
		reflect.TypeOf(uint64(0)),
		reflect.TypeOf(uintptr(0)),
		reflect.TypeOf(float32(0)),
		reflect.TypeOf(float64(0)),
		reflect.TypeOf(complex64(0)),
		reflect.TypeOf(complex128(0)),
		reflect.TypeOf(string("")),
		reflect.TypeOf(false),
		reflect.TypeOf(time.Time{}),
		reflect.TypeOf(time.Duration(0)),
		reflectTypeType,
		reflectValueType,
	)

	if err != nil {
		panic(err)
	}
}

var registered []reflect.Type

// ErrAlreadyRegistered is returned by Register if one or more of the given types has already been registered.
// It is wrapped.
var ErrAlreadyRegistered = errors.New("already registered")

// Register allows an unknown type to be decoded with Type.
func Register(types ...reflect.Type) error {
	var errMsg string
	for _, ty := range types {
		if register(ty) {
			if len(errMsg) > 0 {
				errMsg += ", "
			}
			errMsg += fmt.Sprintf("%v", ty)
			continue
		}
	}

	if errMsg != "" {
		return fmt.Errorf("%w: %v", ErrAlreadyRegistered, errMsg)
	}
	return nil
}

// register registers the type, returning true if it was previously registered.
func register(ty reflect.Type) bool {
	for _, r := range registered {
		if r == ty {
			return true
		}
	}
	registered = append(registered, ty)
	return false
}

// NewType returns a new reflect.Type encodable.
func NewType(config Config) *Type {
	e := &Type{
		typeByID: make(map[ID]reflect.Type),
		idByType: make(map[reflect.Type]ID),
		config:   config,
	}

	for _, t := range registered {
		id := GetID(t, config)

		e.typeByID[id] = t
		e.idByType[t] = id
	}

	return e
}

// Type is an encodable for reflect.Type values.
type Type struct {
	typeByID map[ID]reflect.Type
	idByType map[reflect.Type]ID
	config   Config
}

// Size implements Encodable.
func (e *Type) Size() int { return 16 }

// Type implements Encodable.
func (e *Type) Type() reflect.Type { return reflectValueType }

// Encode implements Encodable.
func (e *Type) Encode(ptr unsafe.Pointer, w io.Writer) error {
	ty := *(*reflect.Type)(ptr)
	id, ok := e.idByType[ty]
	if !ok {
		// This type hasn't been registered. Register it now.
		id = GetID(ty, e.config)

		e.typeByID[id] = ty
		e.idByType[ty] = id
	}

	if e.config&LogTypes != 0 {
		fmt.Fprintf(encio.Warnings, "type %v send with id %v (previously seen: %v)\n", ty.String(), id, ok)
	}

	return encio.Write(id[:], w)
}

// Decode implements Encodable.
// Apart from checking registered types, it also cheks if the type currently in ptr is a match,
// and if so, does nothing.
func (e *Type) Decode(ptr unsafe.Pointer, r io.Reader) error {
	var id ID
	if err := encio.Read(id[:], r); err != nil {
		return err
	}

	ty, ok := e.typeByID[id]
	if ok {
		// Found the type
		*(*reflect.Type)(ptr) = ty
		return nil
	}

	if *(*reflect.Type)(ptr) != nil {
		// No match, check if the existing type is a match.
		ty = *(*reflect.Type)(ptr)
		existing := GetID(ty, e.config)
		if existing == id {
			// A match, add it to our known types and return.

			e.typeByID[id] = ty
			e.idByType[ty] = id

			return nil
		}

		// Neither existing type nor any registered types match strictly. Check loose types.
		var nameMatch reflect.Type
		if existing[0] == id[0] {
			// We have a loose match with the existing.
			if e.config&LooseTyping != 0 {
				return nil
			}
			nameMatch = ty
		} else {
			for rty, rid := range e.idByType {
				if rid[0] == id[0] {
					if e.config&LooseTyping != 0 {
						// We found a loose match with an existing type.
						*(*reflect.Type)(ptr) = rty
						return nil
					}
					nameMatch = rty
				}
			}
		}

		if nameMatch != nil && nameMatch.Name() != "" {
			return encio.NewError(encio.ErrBadType, fmt.Sprintf("unknown type ID %016X. %v loosely matches, but LooseTyping isn't enabled.", id, nameMatch), 0)
		}
	}

	return encio.NewError(encio.ErrBadType, fmt.Sprintf("unknown type ID %016X. Is it registered?", id), 0)
}

// NewValue returns a new reflect.Value encodable.
func NewValue(config Config, src Source) *Value {
	return &Value{
		typeEnc: NewType(config),
		src:     NewCachingSource(src),
		config:  config,
		buff:    make([]byte, 1),
	}
}

// Value is an encodable for reflect.Value values.
type Value struct {
	typeEnc *Type
	config  Config
	src     *CachingSource
	buff    []byte
}

const (
	valueValid = 1 << iota
)

// Size implements Encodable.
func (e *Value) Size() int { return -1 << 31 }

// Type implements Encodable.
func (e *Value) Type() reflect.Type { return reflectValueType }

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
		n := reflect.New(ty).Elem()
		n.Set(v)
		v = n
	}

	enc := e.src.NewEncodable(ty, e.config, nil)

	if err := e.typeEnc.Encode(unsafe.Pointer(&ty), w); err != nil {
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
	if err := e.typeEnc.Decode(unsafe.Pointer(&ty), r); err != nil {
		return err
	}

	enc := e.src.NewEncodable(ty, e.config, nil)

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
		*(*reflect.Value)(ptr) = reflect.Value{}
		return err
	}
	return nil
}
