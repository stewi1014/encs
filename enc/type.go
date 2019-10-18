package enc

import (
	"errors"
	"fmt"
	"hash"
	"hash/crc64"
	"io"
	"reflect"
	"sync"
	"unicode"
	"unicode/utf8"
)

// Basic type constants
var (
	intType        = reflect.TypeOf(int(0))
	int8Type       = reflect.TypeOf(int8(0))
	int16Type      = reflect.TypeOf(int16(0))
	int32Type      = reflect.TypeOf(int32(0))
	int64Type      = reflect.TypeOf(int64(0))
	uintType       = reflect.TypeOf(uint(0))
	uint8Type      = reflect.TypeOf(uint8(0))
	uint16Type     = reflect.TypeOf(uint16(0))
	uint32Type     = reflect.TypeOf(uint32(0))
	uint64Type     = reflect.TypeOf(uint64(0))
	uintptrType    = reflect.TypeOf(uintptr(0))
	float32Type    = reflect.TypeOf(float32(0))
	float64Type    = reflect.TypeOf(float64(0))
	complex64Type  = reflect.TypeOf(complex64(0))
	complex128Type = reflect.TypeOf(complex128(0))
	stringType     = reflect.TypeOf(string(""))
	boolType       = reflect.TypeOf(bool(true))
	interfaceType  = reflect.TypeOf(new(interface{})).Elem()

	invalidType = reflect.TypeOf(nil)
)

// TypeFromKind returns a reflect.Type for the given kind.
// It only functions for simple types;
// int*, uint*, float*, complex*, string and bool
func TypeFromKind(kind reflect.Kind) reflect.Type {
	switch kind {
	case reflect.Int:
		return intType
	case reflect.Int8:
		return int8Type
	case reflect.Int16:
		return int16Type
	case reflect.Int32:
		return int32Type
	case reflect.Int64:
		return int64Type
	case reflect.Uint:
		return uintType
	case reflect.Uint8:
		return uint8Type
	case reflect.Uint16:
		return uint16Type
	case reflect.Uint32:
		return uint32Type
	case reflect.Uint64:
		return uint64Type
	case reflect.Uintptr:
		return uintptrType
	case reflect.Float32:
		return float32Type
	case reflect.Float64:
		return float64Type
	case reflect.Complex64:
		return complex64Type
	case reflect.Complex128:
		return complex128Type
	case reflect.String:
		return stringType
	case reflect.Bool:
		return boolType
	default:
		return invalidType
	}
}

// CanReference returns true if it's possible for parent or its sub-types to contain a reference to element (not element itself).
// It does not follow unsafe.Pointer or uintptr types, and does not count unexported struct fields.
func CanReference(parent, element reflect.Type) bool {
	return canReference(parent, element, make(map[reflect.Type]struct{}))
}

func canReference(parent, elem reflect.Type, seen map[reflect.Type]struct{}) bool {
	if _, has := seen[parent]; has {
		return false
	}
	seen[parent] = struct{}{}
	switch parent.Kind() {
	case reflect.Ptr:
		if parent.Elem() == elem {
			return true
		}
		return canReference(parent.Elem(), elem, seen)

	case reflect.Interface:
		return elem.Implements(parent)

	case reflect.Array, reflect.Slice:
		return canReference(parent.Elem(), elem, seen)

	case reflect.Map:
		if canReference(parent.Key(), elem, seen) {
			return true
		}
		return canReference(parent.Elem(), elem, seen)

	case reflect.Struct:
		for i := 0; i < parent.NumField(); i++ {
			f := parent.Field(i)
			if c, _ := utf8.DecodeRune([]byte(f.Name)); !unicode.IsUpper(c) {
				continue // Don't count unexported struct fields
			}

			if canReference(f.Type, elem, seen) {
				return true
			}
		}
	}

	return false
}

// Type is a system for serilising type information.
type Type interface {
	// Encode encodes a type, writing to w.
	Encode(t reflect.Type, w io.Writer) error

	// Decode decodes a type from r.
	Decode(r io.Reader) (t reflect.Type, err error)

	Size() int
}

// NewTypeRegistry creates a new RegisterTypeEncoder populated with encodable builtin types.
// if hasher is nil, it uses crc64 with ISO polynomial. hasher must be the same for both ends.
func NewTypeRegistry(hasher hash.Hash64) *TypeRegistry {
	if hasher == nil {
		hasher = crc64.New(crc64.MakeTable(crc64.ISO))
	}
	r := &TypeRegistry{
		hasher: hasher,
		byID:   make(map[[8]byte]reflect.Type),
		byType: make(map[reflect.Type][8]byte),
	}

	r.Register(new(interface{}))

	return r
}

// TypeRegistry is a Type that is given concrete types through calls to TypeRegistry(),
// and uses a hash of the types full package import path and name to uniqiely identify it.
// Encode() and Decode() are thread safe, but TypeRegistry() is not; make calls to TypeRegistry() before using.
type TypeRegistry struct {
	hashL  sync.Mutex
	hasher hash.Hash64

	byID   map[[8]byte]reflect.Type
	byType map[reflect.Type][8]byte
}

// Size implements Type.
// TypeRegistry has an undefined length.
func (tr *TypeRegistry) Size() int {
	return -1 << 31
}

var errAlreadyRegistered = errors.New("type is already registered")

// Register notifies the Register of a concrete type. Subsequent calls to Decode will search the types provided here.
// Pointer, Slice, Array and Map types are inferred, and only their component types need to be registered.
// Passing a reflect.Type will register the Type it represents, not the reflect.Type type.
func (tr *TypeRegistry) Register(t interface{}) error {
	if ty, ok := t.(reflect.Type); ok {
		return tr.register(ty)
	}
	return tr.register(reflect.TypeOf(t))
}

func (tr *TypeRegistry) register(ty reflect.Type) error {
	// the idea here is to throw an error if registering failed or was unneccecary.
	switch ty.Kind() {
	case reflect.Struct, reflect.Interface:
		return tr.put(tr.hash(ty), ty)

	case reflect.Ptr, reflect.Slice, reflect.Array:
		if err := tr.register(ty.Elem()); err != nil {
			if errors.Is(err, errAlreadyRegistered) {
				return fmt.Errorf("%v: meta-types do not need to be registered", err)
			}
			return err
		}
		return nil

	case reflect.Map:
		// if one of the component types was already registered, but the other was sucessful,
		// no error should be thrown, as the call to register was important.
		kerr := tr.register(ty.Key())
		if kerr != nil && !errors.Is(kerr, errAlreadyRegistered) {
			return kerr
		}
		// kerr is nil or errAlreadyRegistered

		verr := tr.register(ty.Elem())
		if verr != nil && errors.Is(verr, errAlreadyRegistered) {
			if kerr == nil {
				return nil
			}
			return fmt.Errorf("%v: meta-types do not need to be registered", kerr)
		}
		return verr

	case reflect.Chan, reflect.Func:
		return fmt.Errorf("cannot register type %v", ty)

	}

	return tr.put(tr.hash(ty), ty)
}

func (tr *TypeRegistry) put(h [8]byte, ty reflect.Type) error {
	if existing, ok := tr.byID[h]; ok {
		if existing == ty {
			return errAlreadyRegistered
		}
		return fmt.Errorf("hash colission!!! new type %v has same hash as %v", ty, existing)
	}
	tr.byID[h] = ty
	tr.byType[ty] = h
	return nil
}

// Encode implements TypeResolver
func (tr *TypeRegistry) Encode(t reflect.Type, w io.Writer) error {
	var buff [8]byte
	kind := t.Kind()
	if kind == reflect.Chan || kind == reflect.Func {
		return fmt.Errorf("%v: %v is not encodable", ErrBadType, t)
	}

	switch kind {
	case reflect.Chan, reflect.Func:
		return fmt.Errorf("%v: %v is not encodable", ErrBadType, t)

	case reflect.Ptr:
		buff[0] = uint8(reflect.Ptr)
		if err := write(buff[:1], w); err != nil {
			return err
		}
		return tr.Encode(t.Elem(), w)

	case reflect.Slice:
		buff[0] = uint8(reflect.Slice)
		if err := write(buff[:1], w); err != nil {
			return err
		}
		return tr.Encode(t.Elem(), w)

	case reflect.Array:
		l := uint32(t.Len())
		buff[0] = uint8(reflect.Array)
		buff[1] = uint8(l)
		buff[2] = uint8(l >> 8)
		buff[3] = uint8(l >> 16)
		buff[4] = uint8(l >> 24)
		if err := write(buff[:5], w); err != nil {
			return err
		}
		return tr.Encode(t.Elem(), w)

	case reflect.Map:
		buff[0] = uint8(reflect.Map)
		if err := write(buff[:1], w); err != nil {
			return err
		}
		err := tr.Encode(t.Key(), w)
		if err != nil {
			return err
		}

		return tr.Encode(t.Elem(), w)

	default:

		buff[0] = uint8(kind)
		if err := write(buff[:1], w); err != nil {
			return err
		}

		if kind == reflect.Struct || kind == reflect.Interface {
			var ok bool
			buff, ok = tr.byType[t]
			if !ok {
				return fmt.Errorf("%v: type %v not registered", ErrBadType, t)
			}

			return write(buff[:], w)
		}

		return nil
	}
}

// Decode implements TypeResolver
func (tr *TypeRegistry) Decode(r io.Reader) (reflect.Type, error) {
	var buff [8]byte

	if err := read(buff[:1], r); err != nil {
		return nil, err
	}

	kind := reflect.Kind(buff[0])
	switch kind {
	case reflect.Struct, reflect.Interface:
		if err := read(buff[:], r); err != nil {
			return nil, err
		}

		if ty, ok := tr.byID[buff]; ok {
			return ty, nil
		}
		return nil, fmt.Errorf("%v: unknown type hash, is it registered?", ErrMalformed)

	case reflect.Ptr:
		elem, err := tr.Decode(r)
		if err != nil {
			return nil, err
		}
		return reflect.PtrTo(elem), nil

	case reflect.Slice:
		elem, err := tr.Decode(r)
		if err != nil {
			return nil, err
		}

		return reflect.SliceOf(elem), nil

	case reflect.Array:
		if err := read(buff[:4], r); err != nil {
			return nil, err
		}

		l := uint32(buff[0]) | uint32(buff[1])<<8 | uint32(buff[2])<<16 | uint32(buff[3])<<24

		elem, err := tr.Decode(r)
		if err != nil {
			return nil, err
		}

		if uintptr(l)+elem.Size() > TooBig {
			return nil, fmt.Errorf("%v: given array size is too big, %v bytes", ErrMalformed, uintptr(l)+elem.Size())
		}

		return reflect.ArrayOf(int(l), elem), nil

	case reflect.Map:
		key, err := tr.Decode(r)
		if err != nil {
			return nil, err
		}

		val, err := tr.Decode(r)
		if err != nil {
			return nil, err
		}

		return reflect.MapOf(key, val), nil

	default:
		ty := TypeFromKind(kind)
		if ty == nil || ty.Kind() == reflect.Invalid {
			return ty, fmt.Errorf("%v: received invalid type", ErrMalformed)
		}
		return ty, nil
	}
}

func (tr *TypeRegistry) hash(ty reflect.Type) (out [8]byte) {
	tr.hasher.Reset()
	buff := []byte(tr.name(ty))
	n, err := tr.hasher.Write(buff)
	if err != nil {
		panic(err)
	}
	if n != len(buff) {
		panic(fmt.Errorf("incomplete write of data to hasher, %v byte string but only wrote %v", len(buff), n))
	}
	h := tr.hasher.Sum64()
	out[0] = uint8(h)
	out[1] = uint8(h >> 8)
	out[2] = uint8(h >> 16)
	out[3] = uint8(h >> 24)
	out[4] = uint8(h >> 32)
	out[5] = uint8(h >> 40)
	out[6] = uint8(h >> 48)
	out[7] = uint8(h >> 56)
	return
}

func (tr *TypeRegistry) name(t reflect.Type) string {
	if t.Kind() == reflect.Ptr {
		return "*" + tr.name(t.Elem())
	}
	pkg := t.PkgPath()
	if pkg != "" {
		return pkg + "." + t.Name()
	}
	n := t.Name()
	if n == "" {
		return t.String()
	}
	return n
}
