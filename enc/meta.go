package enc

import (
	"fmt"
	"io"
	"reflect"
	"sort"
	"unicode"
	"unicode/utf8"
	"unsafe"
)

// NewPointer returns a new Pointer Encodable.
func NewPointer(t reflect.Type, config *Config) Encodable {
	if config != nil {
		config = config.copy()
	}
	return newPointer(t, config)
}

func newPointer(t reflect.Type, config *Config) (enc Encodable) {
	if t.Kind() != reflect.Ptr {
		panic(fmt.Errorf("%v: %v is not a pointer", ErrBadType, t))
	}

	if config == nil {
		config = new(Config)
	}

	p := &Pointer{
		ty:   t,
		buff: make([]byte, 1),
	}

	p.r, enc = config.referencer(p)
	p.elem = p.r.newEncodable(t.Elem(), config)
	return
}

// Pointer encodes pointers to concrete types.
type Pointer struct {
	ty   reflect.Type
	r    *referencer
	elem Encodable
	buff []byte
}

const (
	encodedPtr = iota
	nilPtr
)

// Size implements Sized
func (p *Pointer) Size() int {
	return p.elem.Size() + 5
}

// Type implements Encodable
func (p *Pointer) Type() reflect.Type {
	return p.ty
}

// Encode implements Encodable
func (p *Pointer) Encode(ptr unsafe.Pointer, w io.Writer) error {
	if ptr == nil {
		return ErrNilPointer
	}

	return p.r.encodeReference(*(*unsafe.Pointer)(ptr), p.elem, w)
}

// Decode implements Encodable
func (p *Pointer) Decode(ptr unsafe.Pointer, r io.Reader) error {
	if ptr == nil {
		return ErrNilPointer
	}

	return p.r.decodeReference((*unsafe.Pointer)(ptr), p.elem, r)
}

// NewMap returns a new map Encodable
func NewMap(t reflect.Type, config *Config) Map {
	if config != nil {
		config = config.copy()
	}
	return newMap(t, config)
}

func newMap(t reflect.Type, config *Config) Map {
	return Map{
		key:  newEncodable(t.Key(), config),
		val:  newEncodable(t.Elem(), config),
		buff: make([]byte, 4),
		t:    t,
	}
}

// Map is an Encodable for maps
type Map struct {
	key, val Encodable
	buff     []byte
	t        reflect.Type
}

// Size implements Encodable
func (e Map) Size() int {
	return -1 << 31
}

// Type implements Encodable
func (e Map) Type() reflect.Type {
	return e.t
}

// Encode implements Encodable
func (e Map) Encode(ptr unsafe.Pointer, w io.Writer) error {
	if ptr == nil {
		return ErrNilPointer
	}
	v := reflect.NewAt(e.t, ptr).Elem()

	l := uint32(v.Len())
	e.buff[0] = uint8(l)
	e.buff[1] = uint8(l >> 8)
	e.buff[2] = uint8(l >> 16)
	e.buff[3] = uint8(l >> 24)
	if err := write(e.buff, w); err != nil {
		return err
	}

	iter := v.MapRange()
	for iter.Next() {
		err := e.key.Encode(unsafe.Pointer(iter.Key().UnsafeAddr()), w)
		if err != nil {
			return err
		}

		err = e.val.Encode(unsafe.Pointer(iter.Value().UnsafeAddr()), w)
		if err != nil {
			return err
		}
	}

	return nil
}

// Decode implements Encodable
func (e Map) Decode(ptr unsafe.Pointer, r io.Reader) error {
	if ptr == nil {
		return ErrNilPointer
	}
	if err := read(e.buff, r); err != nil {
		return err
	}

	l := uint32(e.buff[0])
	l |= uint32(e.buff[1]) << 8
	l |= uint32(e.buff[2]) << 16
	l |= uint32(e.buff[3]) << 24

	v := reflect.NewAt(e.t, ptr)
	v.Elem().Set(reflect.New(e.t).Elem())

	for i := uint32(0); i < l; i++ {
		nKey := reflect.New(e.key.Type())
		err := e.key.Decode(unsafe.Pointer(nKey.Pointer()), r)
		if err != nil {
			return err
		}

		nVal := reflect.New(e.val.Type())
		err = e.val.Decode(unsafe.Pointer(nVal.Pointer()), r)
		if err != nil {
			return err
		}

		v.SetMapIndex(nKey.Elem(), nVal.Elem())
	}

	return nil
}

// NewInterface returns a new interface Encodable
func NewInterface(t reflect.Type, config *Config) Encodable {
	if config != nil {
		config = config.copy()
	}
	return newInterface(t, config)
}

func newInterface(t reflect.Type, config *Config) (enc Encodable) {
	if t.Kind() != reflect.Interface {
		panic(fmt.Errorf("%v: %v is not an interface", ErrBadType, t))
	}
	if config == nil || config.TypeEncoder == nil {
		panic(fmt.Errorf("Interface Encodable needs non-nil TypeEncoder"))
	}

	i := Interface{
		t:        t,
		encoders: make(map[reflect.Type]Encodable),
		buff:     make([]byte, 1),
		config:   config,
	}

	_, enc = config.referencer(i) // config stores *referencer; just use it's reference.
	return
}

// Interface is an Encodable for interfaces
type Interface struct {
	t        reflect.Type
	config   *Config
	encoders map[reflect.Type]Encodable
	buff     []byte
}

const (
	ifNil = 1 << iota
	ifNonNil
)

// Size implements Encodable
func (e Interface) Size() int {
	return -1 << 31
}

// Type implements Encodable
func (e Interface) Type() reflect.Type {
	return e.t
}

// Encode implements Encodable
func (e Interface) Encode(ptr unsafe.Pointer, w io.Writer) error {
	if ptr == nil {
		return ErrNilPointer
	}

	i := reflect.NewAt(e.t, ptr).Elem()
	if i.IsNil() {
		e.buff[0] = ifNil
		return write(e.buff, w)
	}

	e.buff[0] = ifNonNil
	err := write(e.buff, w)
	if err != nil {
		return err
	}

	elemType := i.Elem().Type()

	err = e.config.TypeEncoder.Encode(elemType, w)
	if err != nil {
		return err
	}

	elemEnc := e.getEncodable(elemType)
	return e.config.r.encodeReference(ptrInterface(ptr).ptr, elemEnc, w)
}

// Decode implements Encodable
func (e Interface) Decode(ptr unsafe.Pointer, r io.Reader) error {
	if ptr == nil {
		return ErrNilPointer
	}
	err := read(e.buff, r)
	if err != nil {
		return err
	}

	i := reflect.NewAt(e.t, ptr).Elem()

	if e.buff[0] == ifNil {
		i.Set(reflect.New(e.t).Elem())
		return nil
	}

	ty, err := e.config.TypeEncoder.Decode(r)
	if err != nil {
		return err
	}

	var eptr unsafe.Pointer
	// re-use existing pointer if possible
	if !i.IsNil() && i.Elem().Type() == ty {
		eptr = unsafe.Pointer(i.Elem().UnsafeAddr())
	}

	enc := e.getEncodable(ty)
	err = e.config.r.decodeReference(&eptr, enc, r)
	if err != nil {
		return err
	}

	new := reflect.NewAt(ty, eptr).Elem()
	i.Set(new)
	return nil
}

func (e Interface) getEncodable(t reflect.Type) Encodable {
	if enc, ok := e.encoders[t]; ok {
		return enc
	}

	enc := newEncodable(t, e.config)
	e.encoders[t] = enc
	return enc
}

// NewSlice returns a new slice Encodable
func NewSlice(t reflect.Type, config *Config) Slice {
	if config != nil {
		config = config.copy()
	}
	return newSlice(t, config)
}

func newSlice(t reflect.Type, config *Config) Slice {
	return Slice{
		elem: newEncodable(t.Elem(), config),
		buff: make([]byte, 4),
	}
}

// Slice is an Encodable for slices
type Slice struct {
	elem Encodable
	buff []byte
}

// Size implemenets Encodable
func (s Slice) Size() int {
	return -1 << 31
}

// Type implements Encodable
func (s Slice) Type() reflect.Type {
	return reflect.SliceOf(s.elem.Type())
}

// Encode implements Encodable
func (s Slice) Encode(ptr unsafe.Pointer, w io.Writer) error {
	if ptr == nil {
		return ErrNilPointer
	}
	sptr := ptrSlice(ptr)
	l := uint32(sptr.len)
	s.buff[0] = uint8(l)
	s.buff[1] = uint8(l >> 8)
	s.buff[2] = uint8(l >> 16)
	s.buff[3] = uint8(l >> 24)
	if err := write(s.buff, w); err != nil {
		return err
	}

	esize := s.elem.Type().Size()

	for i := uint32(0); i < l; i++ {
		eptr := unsafe.Pointer(uintptr(sptr.array) + uintptr(i)*esize)
		err := s.elem.Encode(eptr, w)
		if err != nil {
			return err
		}
	}
	return nil
}

// Decode implemenets Encodable
func (s Slice) Decode(ptr unsafe.Pointer, r io.Reader) error {
	if ptr == nil {
		return ErrNilPointer
	}
	if err := read(s.buff, r); err != nil {
		return err
	}

	l := uint32(s.buff[0])
	l |= uint32(s.buff[1]) << 8
	l |= uint32(s.buff[2]) << 16
	l |= uint32(s.buff[3]) << 24

	sptr := ptrSlice(ptr)
	size := s.elem.Type().Size()

	if uintptr(l)*size > TooBig {
		return fmt.Errorf("%v: size descriptor for slice is too big", ErrMalformed)
	}

	if sptr.array == nil || sptr.cap < int(l) {
		// must allocate
		malloc(size*uintptr(l), &sptr.array)
		sptr.cap = int(l)
	}
	sptr.len = int(l)

	for i := uint32(0); i < l; i++ {
		eptr := unsafe.Pointer(uintptr(sptr.array) + uintptr(i)*size)
		err := s.elem.Decode(eptr, r)
		if err != nil {
			return err
		}
	}

	return nil
}

// NewArray returns a new array Encodable
func NewArray(t reflect.Type, config *Config) Array {
	if config != nil {
		config = config.copy()
	}
	return newArray(t, config)
}

func newArray(t reflect.Type, config *Config) Array {
	return Array{
		elem: newEncodable(t.Elem(), config),
		len:  uintptr(t.Len()),
	}
}

// Array is an Encodable for arrays
type Array struct {
	elem Encodable
	len  uintptr
}

// Size implements Encodable
func (e Array) Size() int {
	s := e.elem.Size()
	if s < 0 {
		return -1 << 31
	}
	return s * int(e.len)
}

// Type implements Encodable
func (e Array) Type() reflect.Type {
	return reflect.ArrayOf(int(e.len), e.elem.Type())
}

// Encode implements Encodable
func (e Array) Encode(ptr unsafe.Pointer, w io.Writer) error {
	if ptr == nil {
		return ErrNilPointer
	}
	esize := e.elem.Type().Size()
	for i := uintptr(0); i < e.len; i++ {
		eptr := unsafe.Pointer(uintptr(ptr) + (i * esize))
		err := e.elem.Encode(eptr, w)
		if err != nil {
			return err
		}
	}
	return nil
}

// Decode implments Encodable
func (e Array) Decode(ptr unsafe.Pointer, r io.Reader) error {
	if ptr == nil {
		return ErrNilPointer
	}
	esize := e.elem.Type().Size()
	for i := uintptr(0); i < e.len; i++ {
		eptr := unsafe.Pointer(uintptr(ptr) + (i * esize))
		err := e.elem.Decode(eptr, r)
		if err != nil {
			return err
		}
	}
	return nil
}

type structMembers []reflect.StructField

func (a structMembers) Len() int           { return len(a) }
func (a structMembers) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a structMembers) Less(i, j int) bool { return a[i].Name < a[j].Name }

// NewStruct returns a new struct Encodable
func NewStruct(t reflect.Type, config *Config) *Struct {
	if config != nil {
		config = config.copy()
	}
	return newStruct(t, config)
}

func newStruct(t reflect.Type, config *Config) *Struct {
	if t.Kind() != reflect.Struct {
		panic(fmt.Errorf("%v: %v is not a struct", ErrBadType, t))
	}

	s := &Struct{
		ty: t,
	}
	n := t.NumField()
	sms := make(structMembers, 0, n)
	for i := 0; i < n; i++ {
		f := t.Field(i)
		if c, _ := utf8.DecodeRune([]byte(f.Name)); unicode.IsUpper(c) || config.IncludeUnexported {
			sms = append(sms, f)
		}
	}

	// struct members are sorted alphabetically. Since there is no coordination of member data,
	// decoders must decode in the same order the encoders wrote.
	// Alphabetically is a pretty platform-independant way of sorting the fields.
	sort.Sort(sms)

	s.members = make([]structMember, len(sms))
	for i := range sms {
		s.members[i] = structMember{
			Encodable: newEncodable(sms[i].Type, config),
			offset:    sms[i].Offset,
		}
	}

	return s
}

// Struct is an Encodable for structs
type Struct struct {
	ty      reflect.Type
	members []structMember
}

type structMember struct {
	Encodable
	ty     reflect.Type
	offset uintptr
}

func (sm structMember) encodeMember(structPtr unsafe.Pointer, w io.Writer) error {
	return sm.Encode(unsafe.Pointer(uintptr(structPtr)+sm.offset), w)
}

func (sm structMember) decodeMember(structPtr unsafe.Pointer, r io.Reader) error {
	return sm.Decode(unsafe.Pointer(uintptr(structPtr)+sm.offset), r)
}

// Size implements Sized
func (e Struct) Size() (size int) {
	for _, member := range e.members {
		msize := member.Size()
		if msize < 0 {
			return -1 << 31
		}
		size += msize
	}
	return
}

// Type implements Encodable
func (e Struct) Type() reflect.Type {
	return e.ty
}

// Encode implemenets Encodable
func (e Struct) Encode(ptr unsafe.Pointer, w io.Writer) error {
	if ptr == nil {
		return ErrNilPointer
	}
	for _, m := range e.members {
		err := m.encodeMember(ptr, w)
		if err != nil {
			return err
		}
	}
	return nil
}

// Decode implemenets Encodable
func (e Struct) Decode(ptr unsafe.Pointer, r io.Reader) error {
	if ptr == nil {
		return ErrNilPointer
	}
	for _, m := range e.members {
		err := m.decodeMember(ptr, r)
		if err != nil {
			return err
		}
	}
	return nil
}
