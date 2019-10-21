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

func newPointer(t reflect.Type, config *Config) (enc Encodable) { // TODO; improve performance in some cases by not using referencer when we can garuntee no self-references *that does not mean only recursive types*.
	if t.Kind() != reflect.Ptr {
		panic(fmt.Errorf("%v: %v is not a pointer", ErrBadType, t))
	}

	if config == nil {
		config = new(Config)
	}

	e := &Pointer{
		ty:   t,
		buff: make([]byte, 1),
	}

	e.r, enc = config.referencer(e)
	e.elem = e.r.newEncodable(t.Elem(), config)
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

func (e *Pointer) String() string {
	if e.r != nil {
		// Not particularly relevant to callers except that when using strings to equality check,
		// the configuration of the resolver is important; it effects the encoded format.
		return fmt.Sprintf("Pointer(resolver at %v){%v}", e.r.Type().String(), e.elem.String())
	}
	return fmt.Sprintf("Pointer{%v}", e.elem.String())
}

// Size implements Sized
func (e *Pointer) Size() int {
	return e.elem.Size() + 5
}

// Type implements Encodable
func (e *Pointer) Type() reflect.Type {
	return e.ty
}

// Encode implements Encodable
func (e *Pointer) Encode(ptr unsafe.Pointer, w io.Writer) error {
	if ptr == nil {
		return ErrNilPointer
	}

	return e.r.encodeReference(*(*unsafe.Pointer)(ptr), e.elem, w)
}

// Decode implements Encodable
func (e *Pointer) Decode(ptr unsafe.Pointer, r io.Reader) error {
	if ptr == nil {
		return ErrNilPointer
	}

	return e.r.decodeReference((*unsafe.Pointer)(ptr), e.elem, r)
}

// NewMap returns a new map Encodable
func NewMap(t reflect.Type, config *Config) *Map {
	if config != nil {
		config = config.copy()
	}
	return newMap(t, config)
}

func newMap(t reflect.Type, config *Config) *Map {
	if t.Kind() != reflect.Map {
		panic(fmt.Errorf("%v: %v is not a map", ErrBadType, t))
	}

	return &Map{
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

func (e *Map) String() string {
	return fmt.Sprintf("Map[%v]{%v}", e.key, e.val)
}

// Size implements Encodable
func (e *Map) Size() int {
	return -1 << 31
}

// Type implements Encodable
func (e *Map) Type() reflect.Type {
	return e.t
}

// Encode implements Encodable
func (e *Map) Encode(ptr unsafe.Pointer, w io.Writer) error {
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
func (e *Map) Decode(ptr unsafe.Pointer, r io.Reader) error {
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
func NewInterface(t reflect.Type, config *Config) Encodable { // TODO; improve performance in some cases by not using a referencer in cases where we can garuntee no self-references.
	if config != nil {
		config = config.copy()
	}
	return newInterface(t, config)
}

func newInterface(t reflect.Type, config *Config) (enc Encodable) {
	if t.Kind() != reflect.Interface {
		panic(fmt.Errorf("%v: %v is not an interface", ErrBadType, t))
	}
	if config == nil || config.Resolver == nil {
		panic(fmt.Errorf("Interface Encodable needs non-nil Resolver"))
	}

	i := &Interface{
		t:        t,
		config:   config,
		encoders: make(map[reflect.Type]Encodable),
		buff:     make([]byte, 1),
	}

	_, enc = config.referencer(i)
	return enc
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

func (e *Interface) String() string {
	return fmt.Sprintf("Interface(Type: %v, %v)", e.t.String(), e.config.String())
}

// Size implements Encodable
func (e *Interface) Size() int {
	return -1 << 31
}

// Type implements Encodable
func (e *Interface) Type() reflect.Type {
	return e.t
}

// Encode implements Encodable
func (e *Interface) Encode(ptr unsafe.Pointer, w io.Writer) error {
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

	err = e.config.Resolver.Encode(elemType, w)
	if err != nil {
		return err
	}

	elemEnc := e.getEncodable(elemType)
	if e.config.r != nil {
		return e.config.r.encodeReference(ptrInterface(ptr).ptr(), elemEnc, w)
	}
	return elemEnc.Encode(ptrInterface(ptr).ptr(), w)
}

// Decode implements Encodable
func (e *Interface) Decode(ptr unsafe.Pointer, r io.Reader) error {
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

	var elemt reflect.Type
	if !i.IsNil() {
		elemt = i.Elem().Type()
	}

	ty, err := e.config.Resolver.Decode(elemt, r)
	if err != nil {
		return err
	}

	// decode, pointing eptr to the decoded value.
	var eptr unsafe.Pointer
	// re-use existing pointer if possible
	if elemt == ty {
		eptr = unsafe.Pointer(ptrInterface(ptr).ptr())
	}

	enc := e.getEncodable(ty)
	if e.config.r != nil {
		// decode into eptr with referencer
		if err := e.config.r.decodeReference(&eptr, enc, r); err != nil {
			return err
		}
	}
	// do our own decoding
	var new reflect.Value
	if eptr == nil {
		// we couldn't re-use our old value
		new = reflect.New(enc.Type())
		eptr = unsafe.Pointer(new.Pointer())
		if err := enc.Decode(eptr, r); err != nil {
			return err
		}
	} else {
		// we could re-use our old value
		new = reflect.NewAt(enc.Type(), eptr)
		if err := enc.Decode(eptr, r); err != nil {
			return err
		}
	}
	i.Set(new.Elem())
	return nil
}

func (e *Interface) getEncodable(t reflect.Type) Encodable {
	if enc, ok := e.encoders[t]; ok {
		return enc
	}

	enc := newEncodable(t, e.config)
	e.encoders[t] = enc
	return enc
}

// NewSlice returns a new slice Encodable
func NewSlice(t reflect.Type, config *Config) *Slice {
	if t.Kind() != reflect.Slice {
		panic(fmt.Errorf("%v: %v is not a slice", ErrBadType, t))
	}
	if config != nil {
		config = config.copy()
	}
	return newSlice(t, config)
}

func newSlice(t reflect.Type, config *Config) *Slice {
	return &Slice{
		elem: newEncodable(t.Elem(), config),
		buff: make([]byte, 4),
	}
}

// Slice is an Encodable for slices
type Slice struct {
	elem Encodable
	buff []byte
}

func (e *Slice) String() string {
	return fmt.Sprintf("Slice[%v]", e.elem)
}

// Size implemenets Encodable
func (e *Slice) Size() int {
	return -1 << 31
}

// Type implements Encodable
func (e *Slice) Type() reflect.Type {
	return reflect.SliceOf(e.elem.Type())
}

// Encode implements Encodable
func (e *Slice) Encode(ptr unsafe.Pointer, w io.Writer) error {
	if ptr == nil {
		return ErrNilPointer
	}
	sptr := ptrSlice(ptr)
	l := uint32(sptr.len)
	e.buff[0] = uint8(l)
	e.buff[1] = uint8(l >> 8)
	e.buff[2] = uint8(l >> 16)
	e.buff[3] = uint8(l >> 24)
	if err := write(e.buff, w); err != nil {
		return err
	}

	esize := e.elem.Type().Size()

	for i := uint32(0); i < l; i++ {
		eptr := unsafe.Pointer(uintptr(sptr.array) + uintptr(i)*esize)
		err := e.elem.Encode(eptr, w)
		if err != nil {
			return err
		}
	}
	return nil
}

// Decode implemenets Encodable
func (e *Slice) Decode(ptr unsafe.Pointer, r io.Reader) error {
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

	sptr := ptrSlice(ptr)
	size := e.elem.Type().Size()

	if uintptr(l)*size > uintptr(TooBig) {
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
		err := e.elem.Decode(eptr, r)
		if err != nil {
			return err
		}
	}

	return nil
}

// NewArray returns a new array Encodable
func NewArray(t reflect.Type, config *Config) *Array {
	if config != nil {
		config = config.copy()
	}
	return newArray(t, config)
}

func newArray(t reflect.Type, config *Config) *Array {
	if t.Kind() != reflect.Array {
		panic(fmt.Errorf("%v: %v is not an array", ErrBadType, t))
	}
	return &Array{
		elem: newEncodable(t.Elem(), config),
		len:  uintptr(t.Len()),
	}
}

// Array is an Encodable for arrays
type Array struct {
	elem Encodable
	len  uintptr
}

func (e *Array) String() string {
	return fmt.Sprintf("Array[%v]{%v}", e.len, e.elem)
}

// Size implements Encodable
func (e *Array) Size() int {
	s := e.elem.Size()
	if s < 0 {
		return -1 << 31
	}
	return s * int(e.len)
}

// Type implements Encodable
func (e *Array) Type() reflect.Type {
	return reflect.ArrayOf(int(e.len), e.elem.Type())
}

// Encode implements Encodable
func (e *Array) Encode(ptr unsafe.Pointer, w io.Writer) error {
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
func (e *Array) Decode(ptr unsafe.Pointer, r io.Reader) error {
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
	offset uintptr
}

func (sm structMember) encodeMember(structPtr unsafe.Pointer, w io.Writer) error {
	return sm.Encode(unsafe.Pointer(uintptr(structPtr)+sm.offset), w)
}

func (sm structMember) decodeMember(structPtr unsafe.Pointer, r io.Reader) error {
	return sm.Decode(unsafe.Pointer(uintptr(structPtr)+sm.offset), r)
}

func (e *Struct) String() string {
	str := "Struct(" + e.ty.String() + "){"

	if len(e.members) == 0 {
		return str + "}"
	}

	str += e.members[0].String()
	for i := 1; i < len(e.members); i++ {
		str += ", " + e.members[i].String()
	}

	return str + "}"
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

// Encode implements Encodable
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

// Decode implements Encodable
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
