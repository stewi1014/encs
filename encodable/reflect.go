package encodable

import (
	"errors"
	"fmt"
	"hash/crc64"
	"io"
	"reflect"
	"strconv"
	"unsafe"

	"github.com/stewi1014/encs/encio"
)

func init() {
	err := Register(
		nil,
		intType,
		int8Type,
		int16Type,
		int32Type,
		int64Type,
		uintType,
		uint8Type,
		uint16Type,
		uint32Type,
		uint64Type,
		uintptrType,
		float32Type,
		float64Type,
		complex64Type,
		complex128Type,
		stringType,
		boolType,
		timeTimeType,
		timeDurationType,
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
add:
	for _, t := range types {
		for _, r := range registered {
			if r == t {
				if len(errMsg) > 0 {
					errMsg += ", "
				}
				errMsg += fmt.Sprintf("%v", t)
				continue add
			}
		}
		registered = append(registered, t)
	}

	if errMsg != "" {
		return fmt.Errorf("%w: %v", ErrAlreadyRegistered, errMsg)
	}
	return nil
}

// GetID returns a unique ID for a given reflect.Type.
func GetID(ty reflect.Type, config Config) (id ID) {
	if ty == nil {
		return ID{0, 0}
	}

	hasher := crc64.New(crc64.MakeTable(crc64.ISO))

	// First half is the loose glob.
	if err := encio.Write([]byte(looseGlob(ty, nil)), hasher); err != nil {
		panic(err)
	}
	id[0] = hasher.Sum64()

	// Second half is the details about the types kind
	hasher.Reset()
	if err := encio.Write([]byte(strictGlob(ty, nil)), hasher); err != nil {
		panic(err)
	}
	id[1] = hasher.Sum64()
	return
}

// ID is a 128 bit ID for a given reflect.Type.
type ID [2]uint64

func (id ID) String() string {
	return fmt.Sprintf("%x-%x", id[0], id[1])
}

// looseGlob returns a glob of data which should loosely identify types.
// int* uint* and float* are considered equal.
// complex64 and complex128 are considered equal.
// structs are considered equal if they have the same name.
// interfaces are considered equal if the have the same name.
func looseGlob(ty reflect.Type, seen map[reflect.Type]int) (str string) {
	if seen == nil {
		seen = make(map[reflect.Type]int)
	}

	switch ty.Kind() {
	case reflect.Invalid:
		str += "invalid"
		return

	case reflect.Bool:
		str += "bool"
		return

	case reflect.Int,
		reflect.Int8,
		reflect.Int16,
		reflect.Int32,
		reflect.Int64,
		reflect.Uint,
		reflect.Uint8,
		reflect.Uint16,
		reflect.Uint32,
		reflect.Uint64,
		reflect.Uintptr,
		reflect.Float32,
		reflect.Float64:
		str += "number"
		return

	case reflect.Complex64,
		reflect.Complex128:
		str += "complex"
		return
	}

	seen[ty]++

	switch ty.Kind() {
	case reflect.Array:
		str += "array" + strconv.Itoa(ty.Len()) + looseGlob(ty.Elem(), seen)
		return

	case reflect.Chan:
		str += "chan" + looseGlob(ty.Elem(), seen)
		return

	case reflect.Func:
		for i := 0; i < ty.NumIn(); i++ {
			str += looseGlob(ty.In(i), seen)
		}

		// Without this, moving the first return value to the last input value
		// would look like an identical function.
		str += "out"

		for i := 0; i < ty.NumOut(); i++ {
			str += looseGlob(ty.Out(i), seen)
		}
		return

	case reflect.Interface:
		if ty.Name() != "" {
			str += ty.PkgPath() + "." + ty.Name()
			return
		}

		// Unlike struct's Field() method, Method() returns in lexographical order,
		// so we don't have to worry about order.
		for i := 0; i < ty.NumMethod(); i++ {
			m := ty.Method(i)
			str += m.Name + looseGlob(m.Type, seen)
		}

		return

	case reflect.Map:
		str += "map[" + looseGlob(ty.Key(), seen) + "]" + looseGlob(ty.Elem(), seen)
		return

	case reflect.Ptr:
		str += "*" + looseGlob(ty.Elem(), seen)
		return

	case reflect.Slice:
		str += "[]" + looseGlob(ty.Elem(), seen)
		return

	case reflect.String:
		str += "string"
		return

	case reflect.Struct:
		if ty.Name() != "" {
			str += ty.PkgPath() + "." + ty.Name()
			return
		}

		fields := structFields(ty)

		for _, field := range fields {
			str += field.Name + looseGlob(field.Type, seen)
		}

		return

	case reflect.UnsafePointer:
		str += "unsafepointer"
		return
	}

	// No kind was matched, a new kind must have been added to golang
	// and the library needs updating.
	panic(encio.NewError(
		encio.ErrBadType,
		fmt.Sprintf("%v is of an unknown kind.", ty),
		0,
	))
}

func strictGlob(ty reflect.Type, seen map[reflect.Type]int) (str string) {
	if seen == nil {
		seen = make(map[reflect.Type]int)
	}

	str += ty.Kind().String()
	str += ty.PkgPath() + ty.Name()

	switch ty.Kind() {
	case reflect.Invalid,
		reflect.Bool,
		reflect.Int,
		reflect.Int8,
		reflect.Int16,
		reflect.Int32,
		reflect.Int64,
		reflect.Uint,
		reflect.Uint8,
		reflect.Uint16,
		reflect.Uint32,
		reflect.Uint64,
		reflect.Uintptr,
		reflect.Float32,
		reflect.Float64,
		reflect.Complex64,
		reflect.Complex128:
		return
	}

	if seen[ty] > 1 {
		return "recursed"
	}
	seen[ty]++

	switch ty.Kind() {
	case reflect.Array:
		str += "array" + strconv.Itoa(ty.Len()) + strictGlob(ty.Elem(), seen)
		return

	case reflect.Chan:
		str += "chan" + strictGlob(ty.Elem(), seen)
		return

	case reflect.Func:
		for i := 0; i < ty.NumIn(); i++ {
			str += strictGlob(ty.In(i), seen)
		}

		// Without this, moving the first return value to the last input value
		// would look like an identical function.
		str += "out"

		for i := 0; i < ty.NumOut(); i++ {
			str += strictGlob(ty.Out(i), seen)
		}
		return

	case reflect.Interface:
		// Unlike struct's Field() method, Method() returns in lexographical order,
		// so we don't have to worry about order.
		for i := 0; i < ty.NumMethod(); i++ {
			m := ty.Method(i)
			str += m.Name + strictGlob(m.Type, seen)
		}

		return

	case reflect.Map:
		str += "map[" + strictGlob(ty.Key(), seen) + "]" + strictGlob(ty.Elem(), seen)
		return

	case reflect.Ptr:
		str += "*" + strictGlob(ty.Elem(), seen)
		return

	case reflect.Slice:
		str += "[]" + strictGlob(ty.Elem(), seen)
		return

	case reflect.String:
		str += "string"
		return

	case reflect.Struct:
		fields := structFields(ty)

		for _, field := range fields {
			str += field.Name + strictGlob(field.Type, seen)
		}

		return

	case reflect.UnsafePointer:
		str += "unsafepointer"
		return
	}

	// No kind was matched, a new kind must have been added to golang
	// and the library needs updating.
	panic(encio.NewError(
		encio.ErrBadType,
		fmt.Sprintf("%v is of an unknown kind.", ty),
		0,
	))
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
	idEnc    encio.UUID
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

	return e.idEnc.EncodeUUID(w, id)
}

// Decode implements Encodable.
// Apart from checking registered types, it also cheks if the type currently in ptr is a match,
// and if so, does nothing.
func (e *Type) Decode(ptr unsafe.Pointer, r io.Reader) error {
	id, err := e.idEnc.DecodeUUID(r)
	if err != nil {
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
