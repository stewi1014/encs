package enc

import (
	"fmt"
	"hash"
	"hash/crc64"
	"io"
	"reflect"
)

// Resolver is a method for encoding types.
type Resolver interface {
	// Encode encodes the type t to w.
	Encode(t reflect.Type, w io.Writer) error

	// Decode reads an encoded type, returning it.
	// Decode must only read the same number of bytes written by Encode().
	//
	// In the case of decoding into interfaces, it might be expected that the received type is the same as the existing type in the interface.
	//
	// I'd like to allow for implementations that only check if types are equal, relying on the expected type as an existing reflect.Type instance.
	// In this case, bool or (bool, error) would be sufficent return values.
	//
	// Other implementations attempt to completely resolve a reflect.Type value from encoded data,
	// in which case expected is an uneccecary argument.
	//
	// In the case of registration-based TypeResolvers, if the interface contains the type that's being sent,
	// it can be fortuitus to register the expected type, as it might not have been registered before.
	//
	// I belive this function format is a good happy medium to allow for these different implementations.
	// On calls from this package, expected will never be nil.
	Decode(expected reflect.Type, r io.Reader) (reflect.Type, error)

	// Size returns the number of bytes the TypeResolver will read and write to the buffer.
	// It can write less, but never more. If length is undefined, Size should return a negative.
	Size() int
}

var errAlreadyRegistered = fmt.Errorf("%v: already registered", ErrBadType)

// NewRegisterResolver returns a new RegisterResolver TypeResolver
func NewRegisterResolver(hasher hash.Hash64) *RegisterResolver {
	if hasher == nil {
		hasher = crc64.New(crc64.MakeTable(crc64.ISO))
	}

	rr := &RegisterResolver{
		hasher:   hasher,
		idByType: make(map[reflect.Type][8]byte),
		typeByID: make(map[[8]byte]reflect.Type),
	}

	for _, T := range builtin {
		if err := rr.Register(T); err != nil {
			panic(err)
		}
	}

	return rr
}

// RegisterResolver is a registration-based TypeResolver.
// All types to be encoded and decoded must be registered with Register(), and a hash of the type is taken to use as an id.
// It aims for speed.
type RegisterResolver struct {
	hasher hash.Hash64

	idByType map[reflect.Type][8]byte
	typeByID map[[8]byte]reflect.Type

	buff [8]byte
}

// Register registers T, &T, []T, and *T if T is a pointer.
func (rr *RegisterResolver) Register(T interface{}) error {
	var ty reflect.Type
	var ok bool
	if ty, ok = T.(reflect.Type); !ok {
		ty = reflect.TypeOf(T)
	}

	if err := rr.put(ty, rr.hash(ty)); err != nil {
		return err
	}

	st := reflect.SliceOf(ty)
	if err := rr.put(st, rr.hash(st)); err != nil && err != errAlreadyRegistered {
		return err
	}

	pt := reflect.PtrTo(ty)
	if err := rr.put(pt, rr.hash(pt)); err != nil && err != errAlreadyRegistered {
		return err
	}

	if ty.Kind() == reflect.Ptr {
		if err := rr.put(ty.Elem(), rr.hash(ty.Elem())); err != nil && err != errAlreadyRegistered {
			return err
		}
	}

	return nil
}

func (rr *RegisterResolver) put(ty reflect.Type, h [8]byte) error {
	if oty, ok := rr.typeByID[h]; ok {
		if oty == ty {
			return errAlreadyRegistered
		}
		return fmt.Errorf("hash colission with %v and %v", oty, ty)
	}
	rr.typeByID[h] = ty
	rr.idByType[ty] = h
	return nil
}

// Size implements TypeResolver
func (rr *RegisterResolver) Size() int {
	return 8
}

// Encode implements TypeResolver
func (rr *RegisterResolver) Encode(ty reflect.Type, w io.Writer) error {
	var ok bool
	if rr.buff, ok = rr.idByType[ty]; ok {
		return write(rr.buff[:], w)
	}

	// ty is not registered.
	// register it now and encode, but return an error.

	if err := rr.put(ty, rr.hash(ty)); err != nil {
		return fmt.Errorf("%v is not registered: %v", ty, err)
	}

	if rr.buff, ok = rr.idByType[ty]; ok {
		err := write(rr.buff[:], w)
		if err != nil {
			return err
		}
		return fmt.Errorf("%v is not registered", ty)
	}
	return fmt.Errorf("registration of %v silently failed", ty)
}

// Decode implements TypeResolver
func (rr *RegisterResolver) Decode(expected reflect.Type, r io.Reader) (reflect.Type, error) {
	if err := read(rr.buff[:], r); err != nil {
		return nil, err
	}

	if ty, ok := rr.typeByID[rr.buff]; ok {
		return ty, nil
	}

	if expected == nil {
		return nil, fmt.Errorf("%v: unknown hash, is the type registered?", ErrBadType)
	}

	h := rr.hash(expected)
	if h != rr.buff {
		return nil, fmt.Errorf("%v: unknown hash, is the type registered?", ErrBadType)
	}

	if err := rr.put(expected, h); err != nil {
		return expected, fmt.Errorf("cannot register expected type: %v", err)
	}

	return expected, nil
}

func (rr *RegisterResolver) hash(ty reflect.Type) (out [8]byte) {
	rr.hasher.Reset()
	buff := []byte(Name(ty))
	n, err := rr.hasher.Write(buff)
	if err != nil {
		panic(err)
	}
	if n != len(buff) {
		panic(fmt.Errorf("incomplete write of data to hasher, %v byte string but only wrote %v", len(buff), n))
	}
	h := rr.hasher.Sum64()
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

/*

// NewTypeRegistry creates a new TypeRegistry.
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
	r.Register(time.Time{})

	return r
}

// TypeRegistry is a Type that is given concrete types through calls to Register(),
// and uses a hash of the types full package import path and name to uniqiely identify it.
// Encode() and Decode() are thread safe, but Register() is not; make calls to Register() before using.
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
*/
