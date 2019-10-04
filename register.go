package gneg

import (
	"errors"
	"fmt"
	"hash"
	"hash/crc64"
	"io"
	"reflect"
	"sync"

	"github.com/stewi1014/gneg/gram"
)

// TypeResolver is a system for serilising type information.
type TypeResolver interface {
	// Encode encodes a type
	Encode(t reflect.Type, w io.Writer) error

	// Decode decodes a type
	// r reads the data written and only the data writen to w in Encode().
	Decode(r io.Reader) (t reflect.Type, err error)
}

var defaultResolver *RegisterResolver

func init() { defaultResolver = NewRegisterResolver(nil) }

// Register registeres the type of t with the default resolver.
func Register(t interface{}) error {
	return defaultResolver.Register(t)
}

// MustRegister is the same as register, only it panics on failure.
func MustRegister(t interface{}) {
	if err := defaultResolver.Register(t); err != nil {
		panic(err)
	}
}

// NewRegisterResolver creates a new RegisterResolver populated with encodable builtin types.
// if hasher is nil, it uses crc64 with ISO polynomial. hasher must be the same for both ends.
func NewRegisterResolver(hasher hash.Hash64) *RegisterResolver {
	if hasher == nil {
		hasher = crc64.New(crc64.MakeTable(crc64.ISO))
	}
	r := &RegisterResolver{
		hasher: hasher,
		byID:   make(map[[8]byte]reflect.Type),
		byType: make(map[reflect.Type][8]byte),
	}
	r.Register(int(0))
	r.Register(int8(0))
	r.Register(int16(0))
	r.Register(int32(0))
	r.Register(int64(0))
	r.Register(uint(0))
	r.Register(uint8(0))
	r.Register(uint16(0))
	r.Register(uint32(0))
	r.Register(uint64(0))
	r.Register(uintptr(0))
	r.Register(float32(0))
	r.Register(float64(0))
	r.Register(complex64(0))
	r.Register(complex128(0))
	r.Register(string(""))
	r.Register(bool(true))
	r.Register(rune(0))
	return r
}

// RegisterResolver is a TypeResolver that is given Concrete types through calls to Register(),
// and uses a hash of the types full package import path and name to uniqiely identify it.
// It is thread safe.
type RegisterResolver struct {
	hashL  sync.Mutex
	hasher hash.Hash64

	mapMutex sync.RWMutex
	byID     map[[8]byte]reflect.Type // map[[8]byte]reflect.Type
	byType   map[reflect.Type][8]byte // map[reflect.Type][8]byte
}

var errAlreadyRegistered = errors.New("type is already registered")

// Register notifies the RegisterResolver of a concrete type. Subsequent calls to Decode will search the types provided here.
// Pointer, Slice, Array and Map types are inferred, and only their component types need to be registered.
func (rr *RegisterResolver) Register(t interface{}) error {
	if ty, ok := t.(reflect.Type); ok {
		return rr.register(ty)
	}
	return rr.register(reflect.TypeOf(t))
}

func (rr *RegisterResolver) register(ty reflect.Type) error {
	switch ty.Kind() {
	case reflect.Ptr, reflect.Slice, reflect.Array:
		return rr.register(ty.Elem())
	case reflect.Map:
		// the idea here is to throw an error if registering failed or was unneccecary.
		// if one of the component types was already registered, but the other was sucessful,
		// no error should be thrown, as the call to register was important.
		kerr := rr.register(ty.Key())
		if kerr != nil && kerr != errAlreadyRegistered {
			return kerr
		}
		verr := rr.register(ty.Elem())
		if kerr == nil {
			// key register was sucessful.
			// if val failed, return, but if it was already registered, ignore
			if verr != errAlreadyRegistered {
				return verr
			}
			return nil
		}
		// key was already registered (other errors caught above).
		// return errors from val.
		return verr

	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Invalid:
		return fmt.Errorf("cannot register type %v", ty)
	}

	return rr.put(rr.hash(ty), ty)
}

func (rr *RegisterResolver) put(h [8]byte, ty reflect.Type) error {
	rr.mapMutex.Lock()
	defer rr.mapMutex.Unlock()

	if existing, ok := rr.byID[h]; ok {
		if existing == ty {
			return errAlreadyRegistered
		}
		return fmt.Errorf("hash colission!!! new type %v has same hash as %v", ty, existing)
	}
	rr.byID[h] = ty
	rr.byType[ty] = h
	return nil
}

const (
	rrHashed = byte(iota)
	rrPtr
	rrSlice
	rrArray
	rrMap
)

// Encode implements TypeResolver
func (rr *RegisterResolver) Encode(t reflect.Type, w io.Writer) error {
	g, write := gram.WriteGram(w)
	err := rr.encode(t, g)
	if err != nil {
		return err
	}

	return write()
}

func (rr *RegisterResolver) encode(ty reflect.Type, g *gram.Gram) error {
	// encoding has a first byte decribing how to decode.
	switch ty.Kind() {
	case reflect.Ptr:
		// pointer type, one byte saying so and the rest for the element type.
		g.WriteBuff(1)[0] = rrPtr
		return rr.encode(ty.Elem(), g)

	case reflect.Slice:
		// slice type, one byte saying so and the rest for the element type.
		g.WriteBuff(1)[0] = rrSlice
		return rr.encode(ty.Elem(), g)

	case reflect.Array:
		// array type, one byte saying so, uint for length and the rest for the element type.
		l := ty.Len()
		g.WriteBuff(1)[0] = rrArray // 1 byte
		g.WriteUint(uint64(l))      // 1-9 bytes
		return rr.encode(ty.Elem(), g)

	case reflect.Map:
		// map type, one byte saying so, length of key type, key type and val type.
		g.WriteBuff(1)[0] = rrMap
		header := gram.WriteSizeHeader(g)
		err := rr.encode(ty.Key(), g)
		if err != nil {
			return err
		}
		header()

		return rr.encode(ty.Elem(), g)

	default:
		// non-inferrable type, must be registered; use hashes, one byte saying so and 8 bytes of hash.
		rr.mapMutex.RLock()
		defer rr.mapMutex.RUnlock()

		if hash, ok := rr.byType[ty]; ok {
			buff := g.WriteBuff(1 + 8)
			buff[0] = rrHashed
			copy(buff[1:], hash[:])
			return nil
		}

		return fmt.Errorf("type %v not registered", ty)
	}
}

// Decode implements TypeResolver
func (rr *RegisterResolver) Decode(r io.Reader) (reflect.Type, error) {
	g, err := gram.ReadGram(r)
	if err != nil {
		return nil, err
	}

	defer g.Close()
	return rr.decode(g)
}

func (rr *RegisterResolver) decode(g *gram.Gram) (reflect.Type, error) {
	encodeType := g.ReadBuff(1)[0]
	switch encodeType {
	case rrHashed:
		var h [8]byte
		c := copy(h[:], g.ReadBuff(8))
		if c != 8 {
			return nil, fmt.Errorf("incomplete read")
		}
		rr.mapMutex.RLock()
		defer rr.mapMutex.RUnlock()
		if ty, ok := rr.byID[h]; ok {
			return ty, nil
		}
		return nil, fmt.Errorf("unknown type")

	case rrPtr:
		elem, err := rr.decode(g)
		if err != nil {
			return nil, err
		}
		return reflect.PtrTo(elem), nil

	case rrSlice:
		elem, err := rr.decode(g)
		if err != nil {
			return nil, err
		}
		return reflect.SliceOf(elem), nil

	case rrArray:
		l := g.ReadUint()
		if l > tooBig {
			return nil, errTooBig
		}
		elem, err := rr.decode(g)
		if err != nil {
			return nil, err
		}
		return reflect.ArrayOf(int(l), elem), nil

	case rrMap:
		keyGram := gram.ReadSizeHeader(g)
		key, err := rr.decode(keyGram)
		if err != nil {
			return nil, err
		}
		val, err := rr.decode(g)
		if err != nil {
			return nil, err
		}
		return reflect.MapOf(key, val), nil

	default:
		return nil, fmt.Errorf("unknown encoding type")
	}
}

func (rr *RegisterResolver) hash(ty reflect.Type) (out [8]byte) {
	rr.hashL.Lock()
	defer rr.hashL.Unlock()

	buff := []byte(rr.name(ty))
	n, err := rr.hasher.Write(buff)
	if err != nil {
		panic(err)
	}
	if n != len(buff) {
		panic(fmt.Errorf("incomplete write of data to hasher, %v byte string but only wrote %v", len(buff), n))
	}
	h := rr.hasher.Sum64()
	rr.hasher.Reset()
	binEnc.PutUint64(out[:], h)
	return
}

func (rr *RegisterResolver) name(t reflect.Type) string {
	if t.Kind() == reflect.Ptr {
		return "*" + rr.name(t.Elem())
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

// NewCachingResolver returns a new caching resolver, caching calls to resolver.
func NewCachingResolver(resolver TypeResolver) *CachingResolver {
	if cr, ok := resolver.(*CachingResolver); ok {
		return cr // no point wrapping ourselves.
	}
	return &CachingResolver{
		resolver: resolver,
		byType:   make(map[reflect.Type]uint16),
		buff:     make([]byte, 2),
	}
}

// CachingResolver wraps a TypeResolver, caching types with 16bit IDs.
// On the first call to Encode, the full type encoding along with the ID is sent.
// After the first instance, only the ID is sent, and lookup tables are used.
//
// It is not thread safe.
type CachingResolver struct {
	resolver TypeResolver

	// decode
	byID []reflect.Type

	// encode
	byType map[reflect.Type]uint16
	lastID uint16

	buff []byte
}

// Encode implements TypeResolver
func (cr *CachingResolver) Encode(t reflect.Type, w io.Writer) error {
	if id, ok := cr.byType[t]; ok {
		n, err := w.Write([]byte{
			byte(id),
			byte(id >> 8),
		})
		if n != 2 {
			if err != nil {
				return err
			}
			return fmt.Errorf("bad write; want 2 bytes but wrote %v", n)
		}
		return err
	}

	cr.lastID++
	id := cr.lastID
	cr.byType[t] = id

	n, err := w.Write([]byte{
		0, 0,
		byte(id),
		byte(id >> 8),
	})
	if n != 4 {
		if err != nil {
			return err
		}
		return fmt.Errorf("bad write; want 4 bytes but wrote %v", n)
	}
	return cr.resolver.Encode(t, w)
}

// Decode implements TypeResolver
func (cr *CachingResolver) Decode(r io.Reader) (reflect.Type, error) {
	n, err := r.Read(cr.buff)
	if n != 2 {
		if err != nil {
			return nil, err
		}
		return nil, fmt.Errorf("short read; want 2 bytes but got %v", n)
	}
	if cr.buff[0] != 0 || cr.buff[1] != 0 {
		id := uint16(cr.buff[0])
		id |= uint16(cr.buff[1]) << 8

		if len(cr.byID) < int(id) {
			return nil, fmt.Errorf("unknown type id %v", id)
		}
		return cr.byID[id-1], nil
	}

	n, err = r.Read(cr.buff)
	if n != 2 {
		if err != nil {
			return nil, err
		}
		return nil, fmt.Errorf("short read; want 2 bytes but got %v", n)
	}

	id := uint16(cr.buff[0])
	id |= uint16(cr.buff[1]) << 8
	ty, err := cr.resolver.Decode(r)
	if err != nil {
		return nil, err
	}
	if int(id) != len(cr.byID)+1 {
		return nil, fmt.Errorf("type infromation out of order")
	}
	cr.byID = append(cr.byID, ty)
	return ty, nil
}
