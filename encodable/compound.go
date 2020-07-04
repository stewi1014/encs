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
		panic(encio.NewError(encio.ErrBadType, fmt.Sprintf("%v is not a pointer", t), 0))
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
	checkPtr(ptr)

	return e.r.encodeReference(*(*unsafe.Pointer)(ptr), e.elem, w)
}

// Decode implements Encodable
func (e *Pointer) Decode(ptr unsafe.Pointer, r io.Reader) error {
	checkPtr(ptr)

	return e.r.decodeReference((*unsafe.Pointer)(ptr), e.elem, r)
}

// NewMap returns a new map Encodable
func NewMap(t reflect.Type, config *Config) *Map {
	return newMap(t, config.genState())
}

func newMap(t reflect.Type, state *state) *Map {
	if t.Kind() != reflect.Map {
		panic(encio.NewError(encio.ErrBadType, fmt.Sprintf("%v is not a map", t), 0))
	}

	return &Map{
		key:     newEncodable(t.Key(), state),
		val:     newEncodable(t.Elem(), state),
		keyBuff: reflect.New(t.Key()).Elem(),
		valBuff: reflect.New(t.Elem()).Elem(),
		t:       t,
	}
}

// Map is an Encodable for maps
type Map struct {
	key, val Encodable
	encio.Uvarint
	keyBuff, valBuff reflect.Value
	t                reflect.Type
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
	checkPtr(ptr)
	v := reflect.NewAt(e.t, ptr).Elem()

	if v.IsNil() {
		return e.Uvarint.Encode(w, 0)
	}

	if err := e.Uvarint.Encode(w, uint32(v.Len()+1)); err != nil {
		return err
	}

	iter := v.MapRange()
	e.keyBuff = reflect.New(v.Type().Key()).Elem()
	e.valBuff = reflect.New(v.Type().Elem()).Elem()
	for iter.Next() {
		e.keyBuff.Set(iter.Key())
		e.valBuff.Set(iter.Value())

		err := e.key.Encode(unsafe.Pointer(e.keyBuff.UnsafeAddr()), w)
		if err != nil {
			return err
		}

		err = e.val.Encode(unsafe.Pointer(e.valBuff.UnsafeAddr()), w)
		if err != nil {
			return err
		}
	}

	return nil
}

// Decode implements Encodable
func (e *Map) Decode(ptr unsafe.Pointer, r io.Reader) error {
	checkPtr(ptr)
	l, err := e.Uvarint.Decode(r)
	if err != nil {
		return err
	}

	m := reflect.NewAt(e.t, ptr).Elem()

	if l == 0 {
		m.Set(reflect.New(e.t).Elem())
		return nil
	}
	l--

	if uintptr(l)*(e.key.Type().Size()+e.val.Type().Size()) > encio.TooBig {
		return encio.NewIOError(encio.ErrMalformed, r, fmt.Sprintf("map size of %v is too big", l), 0)
	}

	v := reflect.MakeMapWithSize(e.t, int(l))

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

	m.Set(v)

	return nil
}

// NewInterface returns a new interface Encodable
func NewInterface(t reflect.Type, config *Config) Encodable { // TODO; improve performance in some cases by not using a referencer in cases where we can garuntee no self-references.
	return newInterface(t, config.genState())
}

func newInterface(t reflect.Type, state *state) (enc Encodable) {
	if t.Kind() != reflect.Interface {
		panic(encio.NewError(encio.ErrBadType, fmt.Sprintf("%v is not an interface", t), 0))
	}
	if state.Resolver == nil {
		panic(encio.NewError(encio.ErrBadConfig, "interface encodables need a resolver to function (config.Resolver is nil)", 0))
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
	checkPtr(ptr)

	i := reflect.NewAt(e.t, ptr).Elem()
	if i.IsNil() {
		e.buff[0] = 0
		return encio.Write(e.buff, w)
	}

	e.buff[0] = 1
	err := encio.Write(e.buff, w)
	if err != nil {
		return err
	}

	elemType := i.Elem().Type()
	elem := reflect.New(elemType).Elem()
	elem.Set(i.Elem())

	err = e.state.Resolver.Encode(elemType, w)
	if err != nil {
		return err
	}

	elemEnc := e.getEncodable(elemType)
	return elemEnc.Encode(unsafe.Pointer(elem.UnsafeAddr()), w)
}

// Decode implements Encodable
func (e *Interface) Decode(ptr unsafe.Pointer, r io.Reader) error {
	checkPtr(ptr)

	if err := encio.Read(e.buff, r); err != nil {
		return err
	}

	i := reflect.NewAt(e.t, ptr).Elem()

	if e.buff[0] == 0 {
		// Nil interface
		i.Set(reflect.New(e.t).Elem())
		return nil
	}

	var elemt reflect.Type
	if !i.IsNil() {
		elemt = i.Elem().Type()
	}

	rty, err := e.state.Resolver.Decode(elemt, r)
	if err != nil {
		return err
	}

	elem := reflect.New(rty).Elem()
	if rty == elemt {
		// The interface already holds a value of the same type as we're receiving.
		// Unfortunately we can't simply go in and modify it directly due to complex interface semantics.
		// What we can do however, is
		elem.Set(i.Elem()) // copy the existing value.

		// Now why would we do that?
		// While we're stuck with the previous allocation, **subsequent encodables do not have to be stuck allocating everything after this**.
		// For instance, if the interface was holding a struct with a slice member then when the slice encodable gets around to do its thing,
		// it will find the backing array pointer and avoid allocating a new one, assuming it's non-nil and cap is large enough.
	}

	enc := e.getEncodable(rty)
	if err := enc.Decode(unsafe.Pointer(elem.UnsafeAddr()), r); err != nil {
		return err
	}

	i.Set(elem)

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
		panic(encio.NewError(encio.ErrBadType, fmt.Sprintf("%v is not a slice", t), 0))
	}

	return &Slice{
		t:    t,
		elem: newEncodable(t.Elem(), state),
	}
}

// Slice is an Encodable for slices
type Slice struct {
	t    reflect.Type
	elem Encodable
	len  encio.Uvarint
}

// String implements Encodable
func (e *Slice) String() string {
	return fmt.Sprintf("[]%v", e.elem)
}

// Size implemenets Encodable
func (e *Slice) Size() int {
	return -1 << 31
}

// Type implements Encodable
func (e *Slice) Type() reflect.Type {
	return reflect.SliceOf(e.elem.Type())
}

// Encode implements Encodable.Encode.
// Encoded 0-len and nil slices both have the effect of setting the decoded slice's
// len and cap to 0. nil-ness of the slice being decoded into is retained.
func (e *Slice) Encode(ptr unsafe.Pointer, w io.Writer) error {
	checkPtr(ptr)

	slice := reflect.NewAt(e.t, ptr).Elem()
	if slice.IsNil() {
		return e.len.Encode(w, 0)
	}

	l := slice.Len()
	if err := e.len.Encode(w, uint32(l+1)); err != nil {
		return err
	}

	for i := 0; i < l; i++ {
		err := e.elem.Encode(unsafe.Pointer(slice.Index(i).UnsafeAddr()), w)
		if err != nil {
			return err
		}
	}
	return nil
}

// Decode implemenets Encodable.
// Encoded 0-len and nil slices both have the effect of setting the decoded slice's
// len and cap to 0. nil-ness of the slice being decoded into is retained.
func (e *Slice) Decode(ptr unsafe.Pointer, r io.Reader) error {
	checkPtr(ptr)
	slice := reflect.NewAt(e.t, ptr).Elem()

	l, err := e.len.Decode(r)
	if err != nil {
		return err
	}
	if l == 0 {
		// Nil slice
		slice.Set(reflect.New(e.t).Elem())
		return nil
	}
	l--

	if uintptr(l)*e.elem.Type().Size() > uintptr(encio.TooBig) {
		return encio.IOError{
			Err:     encio.ErrMalformed,
			Message: fmt.Sprintf("slice of length %v (%v bytes) is too big", l, int(l)*int(e.elem.Type().Size())),
		}
	}

	length := int(l)
	if slice.Cap() < length || slice.Cap() == 0 {
		// Not enough space, allocate
		slice.Set(reflect.MakeSlice(e.t, length, length))
	} else {
		slice.SetLen(length)
	}

	for i := 0; i < length; i++ {
		eptr := unsafe.Pointer(slice.Index(i).UnsafeAddr())
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
		panic(encio.NewError(encio.ErrBadType, fmt.Sprintf("%v is not an Array", t), 0))
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
	checkPtr(ptr)
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
	checkPtr(ptr)
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
		panic(encio.NewError(encio.ErrBadType, fmt.Sprintf("%v is not a struct", t), 0))
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

	// TODO implement Config.StructTag

	// struct members are sorted alphabetically. Since there is no coordination of member data,
	// decoders must decode in the same order the encoders wrote.
	// Alphabetically is a pretty platform-independent way of sorting the fields.
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
	checkPtr(ptr)
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
	checkPtr(ptr)
	for _, m := range e.members {
		err := m.decodeMember(ptr, r)
		if err != nil {
			return err
		}
	}
	return nil
}
