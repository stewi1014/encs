package encodable

import (
	"fmt"
	"io"
	"reflect"
	"sort"
	"unicode"
	"unicode/utf8"
	"unsafe"

	"github.com/stewi1014/encs/encio"
)

// NewPointer returns a new Pointer Encodable.
func NewPointer(t reflect.Type, config *Config) Encodable {
	return newPointer(t, config.genState())
}

func newPointer(t reflect.Type, state *state) (enc Encodable) { // TODO; improve performance in some cases by not using referencer when we can garuntee no self-references *that does not mean only recursive types*.
	if t.Kind() != reflect.Ptr {
		panic(encio.Error{
			Err:     encio.ErrBadType,
			Caller:  "enc.NewPointer",
			Message: fmt.Sprintf("%v is not a pointer", t),
		})
	}

	e := &Pointer{
		ty:   t,
		buff: make([]byte, 1),
	}

	e.r, enc = state.referencer(e)
	e.elem = e.r.newEncodable(t.Elem(), state)
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

// String implements Encodable
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
		return encio.Error{
			Err:    encio.ErrNilPointer,
			Caller: "enc.Pointer.Encode",
		}
	}

	return e.r.encodeReference(*(*unsafe.Pointer)(ptr), e.elem, w)
}

// Decode implements Encodable
func (e *Pointer) Decode(ptr unsafe.Pointer, r io.Reader) error {
	if ptr == nil {
		return encio.Error{
			Err:    encio.ErrNilPointer,
			Caller: "enc.Pointer.Decode",
		}
	}

	return e.r.decodeReference((*unsafe.Pointer)(ptr), e.elem, r)
}

// NewMap returns a new map Encodable
func NewMap(t reflect.Type, config *Config) *Map {
	return newMap(t, config.genState())
}

func newMap(t reflect.Type, state *state) *Map {
	if t.Kind() != reflect.Map {
		panic(encio.Error{
			Err:     encio.ErrBadType,
			Caller:  "enc.NewMap",
			Message: fmt.Sprintf("%v is not a map", t),
		})
	}

	return &Map{
		key:  newEncodable(t.Key(), state),
		val:  newEncodable(t.Elem(), state),
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

// String implements Encodable
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
		return encio.Error{
			Err:    encio.ErrNilPointer,
			Caller: "enc.Map.Encode",
		}
	}
	v := reflect.NewAt(e.t, ptr).Elem()

	l := uint32(v.Len())
	e.buff[0] = uint8(l)
	e.buff[1] = uint8(l >> 8)
	e.buff[2] = uint8(l >> 16)
	e.buff[3] = uint8(l >> 24)
	if err := encio.Write(e.buff, w); err != nil {
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
		return encio.Error{
			Err:    encio.ErrNilPointer,
			Caller: "enc.Map.Decode",
		}
	}
	if err := encio.Read(e.buff, r); err != nil {
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
	return newInterface(t, config.genState())
}

func newInterface(t reflect.Type, state *state) (enc Encodable) {
	if t.Kind() != reflect.Interface {
		panic(encio.Error{
			Err:     encio.ErrBadType,
			Caller:  "enc.NewInterface",
			Message: fmt.Sprintf("%v is not an interface", t),
		})
	}
	if state.Resolver == nil {
		err := encio.Error{
			Err:     encio.ErrBadConfig,
			Caller:  "enc.NewInterface",
			Message: fmt.Sprintf("Interface Encodables need a Resolver to function (config.Resolver is nil)"),
		}

		panic(err)
	}

	i := &Interface{
		t:        t,
		state:    state,
		encoders: make(map[reflect.Type]Encodable),
		buff:     make([]byte, 1),
	}

	_, enc = state.referencer(i)
	return enc
}

// Interface is an Encodable for interfaces
type Interface struct {
	t        reflect.Type
	state    *state
	encoders map[reflect.Type]Encodable
	buff     []byte
}

const (
	ifNil = 1 << iota
	ifNonNil
)

// String implements Encodable
func (e *Interface) String() string {
	return fmt.Sprintf("Interface(Type: %v, %v)", e.t.String(), e.state.String())
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
		return encio.Error{
			Err:    encio.ErrNilPointer,
			Caller: "enc.Interface.Encode",
		}
	}

	i := reflect.NewAt(e.t, ptr).Elem()
	if i.IsNil() {
		e.buff[0] = ifNil
		return encio.Write(e.buff, w)
	}

	e.buff[0] = ifNonNil
	err := encio.Write(e.buff, w)
	if err != nil {
		return err
	}

	elemType := i.Elem().Type()

	err = e.state.Resolver.Encode(elemType, w)
	if err != nil {
		return err
	}

	elemEnc := e.getEncodable(elemType)
	if e.state.r != nil {
		return e.state.r.encodeReference(ptrInterface(ptr).ptr(), elemEnc, w)
	}
	return elemEnc.Encode(ptrInterface(ptr).ptr(), w)
}

// Decode implements Encodable
func (e *Interface) Decode(ptr unsafe.Pointer, r io.Reader) error {
	if ptr == nil {
		return encio.Error{
			Err:    encio.ErrNilPointer,
			Caller: "enc.Interface.Decode",
		}
	}
	err := encio.Read(e.buff, r)
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

	ty, err := e.state.Resolver.Decode(elemt, r)
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
	if e.state.r != nil {
		// decode into eptr with referencer
		if err := e.state.r.decodeReference(&eptr, enc, r); err != nil {
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

	enc := newEncodable(t, e.state)
	e.encoders[t] = enc
	return enc
}

// NewSlice returns a new slice Encodable
func NewSlice(t reflect.Type, config *Config) *Slice {
	return newSlice(t, config.genState())
}

func newSlice(t reflect.Type, state *state) *Slice {
	if t.Kind() != reflect.Slice {
		panic(encio.Error{
			Err:     encio.ErrBadType,
			Caller:  "enc.NewSlice",
			Message: fmt.Sprintf("%v is not a slice", t),
		})
	}
	return &Slice{
		elem: newEncodable(t.Elem(), state),
		buff: make([]byte, 4),
	}
}

// Slice is an Encodable for slices
type Slice struct {
	elem Encodable
	buff []byte
}

// String implements Encodable
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
		return encio.Error{
			Err:    encio.ErrNilPointer,
			Caller: "enc.Slice.Encode",
		}
	}
	sptr := ptrSlice(ptr)
	l := uint32(sptr.len)
	e.buff[0] = uint8(l)
	e.buff[1] = uint8(l >> 8)
	e.buff[2] = uint8(l >> 16)
	e.buff[3] = uint8(l >> 24)
	if err := encio.Write(e.buff, w); err != nil {
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
		return encio.Error{
			Err:    encio.ErrNilPointer,
			Caller: "enc.Slice.Decode",
		}
	}
	if err := encio.Read(e.buff, r); err != nil {
		return err
	}

	l := uint32(e.buff[0])
	l |= uint32(e.buff[1]) << 8
	l |= uint32(e.buff[2]) << 16
	l |= uint32(e.buff[3]) << 24

	sptr := ptrSlice(ptr)
	size := e.elem.Type().Size()

	if uintptr(l)*size > uintptr(encio.TooBig) {
		return encio.IOError{
			Err:     encio.ErrMalformed,
			Message: fmt.Sprintf("slice of length %v (%v bytes) is too big", l, int(l)*int(size)),
		}
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
	return newArray(t, config.genState())
}

func newArray(t reflect.Type, state *state) *Array {
	if t.Kind() != reflect.Array {
		panic(encio.Error{
			Err:     encio.ErrBadType,
			Caller:  "enc.NewArray",
			Message: fmt.Sprintf("%v is not an Array", t),
		})
	}
	return &Array{
		elem: newEncodable(t.Elem(), state),
		len:  uintptr(t.Len()),
	}
}

// Array is an Encodable for arrays
type Array struct {
	elem Encodable
	len  uintptr
}

// String implements Encodable
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
		return encio.Error{
			Err:    encio.ErrNilPointer,
			Caller: "enc.Array.Encode",
		}
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
		return encio.Error{
			Err:    encio.ErrNilPointer,
			Caller: "enc.Array.Decode",
		}
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
	return newStruct(t, config.genState())
}

func newStruct(t reflect.Type, state *state) *Struct {
	if t.Kind() != reflect.Struct {
		panic(encio.Error{
			Err:     encio.ErrBadType,
			Caller:  "enc.NewStruct",
			Message: fmt.Sprintf("%v is not a struct", t),
		})
	}

	s := &Struct{
		ty: t,
	}
	n := t.NumField()
	sms := make(structMembers, 0, n)
	for i := 0; i < n; i++ {
		f := t.Field(i)
		if c, _ := utf8.DecodeRune([]byte(f.Name)); unicode.IsUpper(c) || state.IncludeUnexported {
			sms = append(sms, f)
		}
	}

	//TODO implement Config.StructTag

	// struct members are sorted alphabetically. Since there is no coordination of member data,
	// decoders must decode in the same order the encoders wrote.
	// Alphabetically is a pretty platform-independant way of sorting the fields.
	sort.Sort(sms)

	s.members = make([]structMember, len(sms))
	for i := range sms {
		s.members[i] = structMember{
			Encodable: newEncodable(sms[i].Type, state),
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

// String implements Encodable
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
		return encio.Error{
			Err:    encio.ErrNilPointer,
			Caller: "enc.Struct.Encode",
		}
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
		return encio.Error{
			Err:    encio.ErrNilPointer,
			Caller: "enc.Struct.Decode",
		}
	}
	for _, m := range e.members {
		err := m.decodeMember(ptr, r)
		if err != nil {
			return err
		}
	}
	return nil
}