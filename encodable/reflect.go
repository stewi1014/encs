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

// ID is a 128 bit ID for a given reflect.Type.
type ID [2]uint64

// GetID returns a unique ID for a given reflect.Type.
func GetID(t reflect.Type, config Config) (id ID) {
	hasher := crc64.New(crc64.MakeTable(crc64.ISO))

	// First half is the hashed name
	if err := encio.Write([]byte(name(t)), hasher); err != nil {
		panic(err)
	}
	id[0] = hasher.Sum64()

	// Second half is the details about the types kind
	hasher.Reset()
	if err := encio.Write([]byte(glob(t, config, nil)), hasher); err != nil {
		panic(err)
	}
	id[1] = hasher.Sum64()
	return
}

// name returns the name of a named type. For unnamed types it returns their normal.
func name(t reflect.Type) string {
	if t.Kind() == reflect.Ptr {
		return "*" + name(t.Elem())
	}
	pkg := t.PkgPath()
	if pkg != "" {
		return pkg + "." + t.Name()
	}
	n := t.Name()
	if n != "" {
		return n
	}
	return t.String()
}

// glob returns a glob of data containing as much unique information about the type as possible.
// Used for comparisons; only the hash of this is used.
func glob(t reflect.Type, config Config, seen map[reflect.Type]int) (g string) {
	if seen == nil {
		seen = make(map[reflect.Type]int)
	}

	g += t.Kind().String()

	// We allow ourselves to follow a recursive type once
	if seen[t] > 1 {
		return
	}
	seen[t]++

	switch t.Kind() {
	// Simple types, no extra information to add.
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
		reflect.Complex128,
		reflect.String,
		reflect.UnsafePointer:
		return

	// Compound types. These types are made of other types. Check those too.

	case reflect.Array:
		g += strconv.Itoa(t.Len()) + glob(t.Elem(), config, seen)
		return

	case reflect.Chan:
		g += strconv.Itoa(int(t.ChanDir())) + glob(t.Elem(), config, seen)
		return

	case reflect.Func:
		for i := 0; i < t.NumIn(); i++ {
			g += glob(t.In(i), config, seen)
		}

		// Without this, moving the first return value to the last input value
		// would look like an identical function.
		g += "out"

		for i := 0; i < t.NumOut(); i++ {
			g += glob(t.Out(i), config, seen)
		}

		return

	case reflect.Interface:
		if t.Name() != "" && config&LooseTyping > 0 {
			// We're a named interface, and we've been asked to use loose typing.
			g += "skipped"
			return
		}

		// Unlike struct's Field() method, Method() returns in lexographical order,
		// so we don't have to worry about order.
		for i := 0; i < t.NumMethod(); i++ {
			m := t.Method(i)
			g += m.Name + glob(m.Type, config, seen)
		}

		return

	case reflect.Map:
		g += glob(t.Key(), config, seen) + glob(t.Elem(), config, seen)
		return

	case reflect.Ptr, reflect.Slice:
		g += glob(t.Elem(), config, seen)
		return

	case reflect.Struct:
		if t.Name() != "" && config&LooseTyping > 0 {
			// We're a named struct and we've been asked to use loose typing
			g += "skipped"
			return
		}

		fields := structFields(t, StructTag)

		for _, field := range fields {
			g += field.Name + glob(field.Type, config, seen)
		}

		return
	}

	// No kind was matched, a new kind must have been added to golang
	// and the library needs updating.
	panic(encio.NewError(
		encio.ErrBadType,
		fmt.Sprintf("%v is of an unknown kind.", t),
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
	if !ok && *(*reflect.Type)(ptr) != nil {
		// No match, check if the existing type is a match.
		ty = *(*reflect.Type)(ptr)
		existing := GetID(ty, e.config)
		if existing == id {
			// A match, add it to our known types and return.

			e.typeByID[id] = ty
			e.idByType[ty] = id

			return nil
		}

		// Neither existing type nor any registered types match. Try search by name so we can shor a better error message.
		var nameMatch reflect.Type
		if existing[0] == id[0] {
			nameMatch = ty
		} else {
			for rty, rid := range e.idByType {
				if rid[0] == id[0] {
					nameMatch = rty
				}
			}
		}

		if nameMatch != nil && nameMatch.Name() != "" {
			return encio.NewError(encio.ErrBadType, fmt.Sprintf("unknown type ID %016X. %v has the same name, but its signature doesn't match. Is it different on the remote?", id, nameMatch), 0)
		}
		return encio.NewError(encio.ErrBadType, fmt.Sprintf("unknown type ID %016X. Is it registered?", id), 0)
	}

	*(*reflect.Type)(ptr) = ty
	return nil
}

// NewValue returns a new reflect.Value encodable.
func NewValue(config Config, src Source) *Value {
	return &Value{
		typeEnc: NewType(config),
		src:     NewCachingSource(src),
		config:  config,
	}
}

// Value is an encodable for reflect.Value values.
type Value struct {
	typeEnc *Type
	config  Config
	src     *CachingSource
}

// Size implements Encodable.
func (e *Value) Size() int { return -1 << 31 }

// Type implements Encodable.
func (e *Value) Type() reflect.Type { return reflectValueType }

// Encode implements Encodable.
func (e *Value) Encode(ptr unsafe.Pointer, w io.Writer) error {
	checkPtr(ptr)
	v := (*reflect.Value)(ptr)
	ty := v.Type()

	enc := e.src.NewEncodable(ty, e.config)

	if err := e.typeEnc.Encode(unsafe.Pointer(&ty), w); err != nil {
		return err
	}

	if v.CanAddr() {
		return enc.Encode(unsafe.Pointer(v.UnsafeAddr()), w)
	}
	n := reflect.New(ty).Elem()
	n.Set(*v)
	return enc.Encode(unsafe.Pointer(n.UnsafeAddr()), w)
}

// Decode implements Encodable.
func (e *Value) Decode(ptr unsafe.Pointer, r io.Reader) error {
	checkPtr(ptr)
	v := (*reflect.Value)(ptr)

	var ty reflect.Type
	if err := e.typeEnc.Decode(unsafe.Pointer(&ty), r); err != nil {
		return err
	}

	enc := e.src.NewEncodable(ty, e.config)

	if !v.CanAddr() {
		n := reflect.New(ty).Elem()
		if v.IsValid() && ty == v.Type() {
			n.Set(*v)
		}
		v = &n
	}

	elem := unsafe.Pointer(v.UnsafeAddr())

	if err := enc.Decode(elem, r); err != nil {
		// I no longer trust whatever is in the value
		*v = reflect.ValueOf(nil)
		return err
	}

	*v = reflect.ValueOf(reflect.NewAt(ty, elem))
	return nil
}
