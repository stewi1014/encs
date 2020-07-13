package encodable

import (
	"fmt"
	"hash/crc32"
	"io"
	"reflect"
	"sort"
	"strconv"
	"unicode"
	"unsafe"

	"github.com/stewi1014/encs/encio"
)

// NewPointer returns a new pointer Encodable.
func NewPointer(ty reflect.Type, config Config, src Source) Encodable {
	if ty.Kind() != reflect.Ptr {
		panic(encio.NewError(encio.ErrBadType, fmt.Sprintf("%v is not a pointer", ty), 0))
	}

	return &Pointer{
		buff: make([]byte, 1),
		ty:   ty,
		elem: src.NewEncodable(ty.Elem(), config),
	}
}

const (
	nilPointer = -1
)

// Pointer encodes pointers to concrete types.
type Pointer struct {
	ty   reflect.Type
	elem Encodable
	buff []byte
}

// Size implements Sized.
func (e *Pointer) Size() int {
	return e.elem.Size() + 1
}

// Type implements Encodable.
func (e *Pointer) Type() reflect.Type {
	return e.ty
}

// Encode implements Encodable.
func (e *Pointer) Encode(ptr unsafe.Pointer, w io.Writer) error {
	checkPtr(ptr)
	if *(*unsafe.Pointer)(ptr) == nil {
		// Nil pointer
		e.buff[0] = 1
		return encio.Write(e.buff, w)
	}

	e.buff[0] = 0
	if err := encio.Write(e.buff, w); err != nil {
		return err
	}

	return e.elem.Encode(*(*unsafe.Pointer)(ptr), w)
}

// Decode implements Encodable.
func (e *Pointer) Decode(ptr unsafe.Pointer, r io.Reader) error {
	checkPtr(ptr)

	err := encio.Read(e.buff, r)
	if err != nil {
		return err
	}

	if e.buff[0] != 0 {
		// Nil pointer
		v := reflect.NewAt(e.ty, ptr).Elem()
		v.Set(reflect.New(e.ty).Elem())
		return nil
	}

	if *(*unsafe.Pointer)(ptr) == nil {
		reflect.NewAt(e.ty, ptr).Elem().Set(reflect.New(e.ty.Elem()))
	}

	eptr := *(*unsafe.Pointer)(ptr)
	if err := e.elem.Decode(eptr, r); err != nil {
		return err
	}

	return nil
}

// NewMap returns a new map Encodable.
func NewMap(ty reflect.Type, config Config, src Source) *Map {
	if ty.Kind() != reflect.Map {
		panic(encio.NewError(encio.ErrBadType, fmt.Sprintf("%v is not a map", ty), 0))
	}

	return &Map{
		key: src.NewEncodable(ty.Key(), config),
		val: src.NewEncodable(ty.Elem(), config),
		t:   ty,
	}
}

// Map is an Encodable for maps
type Map struct {
	key, val Encodable
	len      encio.Int
	t        reflect.Type
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
		return e.len.EncodeInt32(w, nilPointer)
	}

	if err := e.len.EncodeInt32(w, int32(v.Len())); err != nil {
		return err
	}

	key := reflect.New(e.t.Key()).Elem()
	val := reflect.New(e.t.Elem()).Elem()

	iter := v.MapRange()
	for iter.Next() {
		key.Set(iter.Key())
		val.Set(iter.Value())

		err := e.key.Encode(unsafe.Pointer(key.UnsafeAddr()), w)
		if err != nil {
			return err
		}

		err = e.val.Encode(unsafe.Pointer(val.UnsafeAddr()), w)
		if err != nil {
			return err
		}
	}

	return nil
}

// Decode implements Encodable
func (e *Map) Decode(ptr unsafe.Pointer, r io.Reader) error {
	checkPtr(ptr)
	l, err := e.len.DecodeInt32(r)
	if err != nil {
		return err
	}

	m := reflect.NewAt(e.t, ptr).Elem()

	if l == nilPointer {
		m.Set(reflect.New(e.t).Elem())
		return nil
	}

	if uintptr(l)*(e.key.Type().Size()+e.val.Type().Size()) > encio.TooBig {
		return encio.NewIOError(encio.ErrMalformed, r, fmt.Sprintf("map size of %v is too big", l), 0)
	}

	v := reflect.MakeMapWithSize(e.t, int(l))
	m.Set(v)

	for i := int32(0); i < l; i++ {
		nKey := reflect.New(e.key.Type()).Elem()
		err := e.key.Decode(unsafe.Pointer(nKey.UnsafeAddr()), r)
		if err != nil {
			return err
		}

		nVal := reflect.New(e.val.Type()).Elem()
		err = e.val.Decode(unsafe.Pointer(nVal.UnsafeAddr()), r)
		if err != nil {
			return err
		}

		v.SetMapIndex(nKey, nVal)
	}

	return nil
}

// NewInterface returns a new interface Encodable
func NewInterface(ty reflect.Type, config Config, src Source) *Interface {
	if ty.Kind() != reflect.Interface {
		panic(encio.NewError(encio.ErrBadType, fmt.Sprintf("%v is not an interface", ty), 0))
	}

	e := &Interface{
		ty:       ty,
		source:   src,
		encoders: make(map[reflect.Type]Encodable),
		typeEnc:  NewType(config),
		buff:     make([]byte, 1),
	}

	return e
}

// Interface is an Encodable for interfaces
type Interface struct {
	ty       reflect.Type
	source   Source
	config   Config
	encoders map[reflect.Type]Encodable
	typeEnc  *Type
	buff     []byte
}

// Size implements Encodable
func (e *Interface) Size() int {
	return -1 << 31
}

// Type implements Encodable
func (e *Interface) Type() reflect.Type {
	return e.ty
}

// Encode implements Encodable
func (e *Interface) Encode(ptr unsafe.Pointer, w io.Writer) error {
	checkPtr(ptr)

	i := reflect.NewAt(e.ty, ptr).Elem()
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

	err = e.typeEnc.Encode(unsafe.Pointer(&elemType), w)
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

	i := reflect.NewAt(e.ty, ptr).Elem()

	if e.buff[0] == 0 {
		// Nil interface
		i.Set(reflect.New(e.ty).Elem())
		return nil
	}

	var elemt reflect.Type
	if !i.IsNil() {
		elemt = i.Elem().Type()
	}

	var rty reflect.Type
	err := e.typeEnc.Decode(unsafe.Pointer(&rty), r)
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

	if !rty.Implements(e.ty) {
		// If loose typing is enabled, then there's a possibility the decoded type doesn't implement the interface.
		return encio.NewError(
			encio.ErrBadType,
			fmt.Sprintf("%v was sent to us inside the %v interface, but %v does not implement %v! The types must be different, has a function been added to the interface?", rty, i.Type(), rty, i.Type()),
			0,
		)
	}

	i.Set(elem)

	return nil
}

func (e *Interface) getEncodable(ty reflect.Type) Encodable {
	enc, ok := e.encoders[ty]
	if ok {
		return enc
	}

	enc = e.source.NewEncodable(ty, e.config)
	e.encoders[ty] = enc
	return enc
}

// NewSlice returns a new slice Encodable
func NewSlice(ty reflect.Type, config Config, src Source) *Slice {
	if ty.Kind() != reflect.Slice {
		panic(encio.NewError(encio.ErrBadType, fmt.Sprintf("%v is not a slice", ty), 0))
	}

	return &Slice{
		t:    ty,
		elem: src.NewEncodable(ty.Elem(), config),
	}
}

// Slice is an Encodable for slices
type Slice struct {
	t    reflect.Type
	elem Encodable
	len  encio.Int
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
		return e.len.EncodeInt32(w, nilPointer)
	}

	l := slice.Len()
	if err := e.len.EncodeInt32(w, int32(l)); err != nil {
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

	l, err := e.len.DecodeInt32(r)
	if err != nil {
		return err
	}
	if l == nilPointer {
		// Nil slice
		slice.Set(reflect.New(e.t).Elem())
		return nil
	}

	if uintptr(l)*(e.elem.Type().Size()) > uintptr(encio.TooBig) {
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
func NewArray(ty reflect.Type, config Config, src Source) *Array {
	if ty.Kind() != reflect.Array {
		panic(encio.NewError(encio.ErrBadType, fmt.Sprintf("%v is not an Array", ty), 0))
	}

	return &Array{
		len:  uintptr(ty.Len()),
		size: ty.Elem().Size(),
		elem: src.NewEncodable(ty.Elem(), config),
	}
}

// Array is an Encodable for arrays
type Array struct {
	elem Encodable
	len  uintptr
	size uintptr
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
	for i := uintptr(0); i < e.len; i++ {
		eptr := unsafe.Pointer(uintptr(ptr) + (i * e.size))
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
	for i := uintptr(0); i < e.len; i++ {
		eptr := unsafe.Pointer(uintptr(ptr) + (i * e.size))
		err := e.elem.Decode(eptr, r)
		if err != nil {
			return err
		}
	}
	return nil
}

// NewStruct return a new struct Encodable.
// It creates a StructLoose or StructStrict Encodable depending on if LooseTyping is set in config.
func NewStruct(ty reflect.Type, config Config, src Source) Encodable {
	if config&LooseTyping > 0 {
		return NewStructLoose(ty, config, src)
	}
	return NewStructStrict(ty, config, src)
}

type structMembers []reflect.StructField

func (a structMembers) Len() int           { return len(a) }
func (a structMembers) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a structMembers) Less(i, j int) bool { return a[i].Name < a[j].Name }

func structFields(ty reflect.Type, structTag string) []reflect.StructField {
	fields := make(structMembers, 0, ty.NumField())
	for i := 0; i < ty.NumField(); i++ {
		field := ty.Field(i)

		tagStr, tagged := field.Tag.Lookup(structTag)
		var tag bool
		if tagged {
			parsed, err := strconv.ParseBool(tagStr)
			if err != nil {
				fmt.Fprintf(encio.Warnings, "%v (decoding struct tag in %v)", err, ty.String())
				tagged = false
			} else {
				tag = parsed
			}
		}

		if tagged && !tag {
			// Tag says skip
			continue
		}

		if !tagged && !unicode.IsUpper([]rune(field.Name)[0]) {
			// Not tagged and not exported
			continue
		}

		fields = append(fields, field)
	}

	// Some optimisations and use cases depend on the order of fields remaining the same between systems, so we sort alphabetically.
	sort.Sort(fields)
	return fields
}

// NewStructLoose returns a new struct Encodable.
func NewStructLoose(ty reflect.Type, config Config, src Source) *StructLoose {
	if ty.Kind() != reflect.Struct {
		panic(encio.NewError(encio.ErrBadType, fmt.Sprintf("%v is not a struct", ty), 0))
	}

	e := &StructLoose{
		ty: ty,
	}

	// Take a hash of each field name, generate an ID from it,
	// and populate fields with the id, Encodable and offset for the field.
	fields := structFields(ty, StructTag)
	hasher := crc32.NewIEEE()
	for _, field := range fields {
		if err := encio.Write([]byte(field.Name), hasher); err != nil {
			panic(err) // hasher should never fail to write.
		}

		f := looseField{
			offset: field.Offset,
			id:     hasher.Sum32(),
			enc:    src.NewEncodable(field.Type, config),
		}

		hasher.Reset()

		for _, existing := range e.fields {
			if existing.id == f.id {
				panic(encio.NewError(
					encio.ErrHashColission,
					fmt.Sprintf("struct field %v with type %v has same hash as previous field with type %v in struct %v", field.Name, field.Type.String(), existing.enc.Type(), ty.String()),
					0,
				))
			}
		}

		e.fields = append(e.fields, &f)
	}

	return e
}

// StructLoose is an Encodable for structs.
// It encodes fields by a generated ID, and maps fields with the same name.
// Exported fields can be ignored using the tag `encs:"false"`, and
// unexported fields can be included with the tag `encs:"true"`.
type StructLoose struct {
	ty      reflect.Type
	fields  []*looseField
	uintEnc encio.Uint
}

type looseField struct {
	id     uint32
	offset uintptr
	enc    Encodable
}

// Size implements Encodable.
func (e *StructLoose) Size() int {
	size := len(e.fields) * 4
	for _, field := range e.fields {
		fsize := field.enc.Size()
		if fsize < 0 {
			return -1 << 31
		}
		size += fsize
	}
	return size
}

// Type implements Encodable.
func (e *StructLoose) Type() reflect.Type { return e.ty }

// Encode implements Encodable.
func (e *StructLoose) Encode(ptr unsafe.Pointer, w io.Writer) error {
	if err := e.uintEnc.EncodeUint32(w, uint32(len(e.fields))); err != nil {
		return err
	}

	for _, field := range e.fields {
		if err := e.uintEnc.EncodeUint32(w, field.id); err != nil {
			return err
		}

		if err := field.enc.Encode(unsafe.Pointer(uintptr(ptr)+field.offset), w); err != nil {
			return err
		}
	}

	return nil
}

// Decode implements Encodable.
// Fields that are not received are set to their zero value.
// Fields sent that do not exist locally are ignored.
func (e *StructLoose) Decode(ptr unsafe.Pointer, r io.Reader) error {
	l, err := e.uintEnc.DecodeUint32(r)
	if err != nil {
		return err
	}
	if uintptr(l)*4 > encio.TooBig { // Don't include size of encoded struct member, way too much work here.
		return encio.NewError(encio.ErrMalformed, fmt.Sprintf("%v struct fields is too many to decode", l), 0)
	}

	var i int
	for parsed := uint32(0); parsed < l; parsed++ {
		id, err := e.uintEnc.DecodeUint32(r)
		if err != nil {
			return err
		}

		// Move next field to the current index. At the end we have all unfetched fields
		// at the end of the slice above index.
		for j := i; j < len(e.fields); j++ {
			if e.fields[j].id == id {
				tmp := e.fields[i]
				e.fields[i] = e.fields[j]
				e.fields[j] = tmp
				goto foundField
			}
		}
		// local field not found
		continue

	foundField:

		if err := e.fields[i].enc.Decode(unsafe.Pointer(uintptr(ptr)+e.fields[i].offset), r); err != nil {
			return err
		}

		i++
	}

	// Set undecoded fields to zero value.
	for _, field := range e.fields[i:] {
		// what a mouthful
		reflect.NewAt(field.enc.Type(), unsafe.Pointer(uintptr(ptr)+field.offset)).Elem().Set(reflect.New(field.enc.Type()).Elem())
	}

	return nil
}

// NewStructStrict returns a new struct Encodable.
func NewStructStrict(ty reflect.Type, config Config, src Source) *StructStrict {
	if ty.Kind() != reflect.Struct {
		panic(encio.NewError(encio.ErrBadType, fmt.Sprintf("%v is not a struct", ty), 0))
	}

	fields := structFields(ty, StructTag)
	s := &StructStrict{
		ty:     ty,
		fields: make([]strictField, len(fields)),
	}

	for i := range fields {
		s.fields[i].offset = fields[i].Offset
		s.fields[i].enc = src.NewEncodable(fields[i].Type, config)
	}

	return s
}

// StructStrict is an Encodable for structs.
// It encodes fields in a determenistic order.
// Exported fields can be ignored using the tag `encs:"false"`, and
// unexported fields can be included with the tag `encs:"true"`.
type StructStrict struct {
	ty     reflect.Type
	fields []strictField
}

type strictField struct {
	offset uintptr
	enc    Encodable
}

// Size implements Sized
func (e StructStrict) Size() (size int) {
	for _, member := range e.fields {
		msize := member.enc.Size()
		if msize < 0 {
			return -1 << 31
		}
		size += msize
	}
	return
}

// Type implements Encodable
func (e StructStrict) Type() reflect.Type { return e.ty }

// Encode implements Encodable
func (e StructStrict) Encode(ptr unsafe.Pointer, w io.Writer) error {
	checkPtr(ptr)
	for _, m := range e.fields {
		if err := m.enc.Encode(unsafe.Pointer(uintptr(ptr)+m.offset), w); err != nil {
			return err
		}
	}
	return nil
}

// Decode implements Encodable
func (e StructStrict) Decode(ptr unsafe.Pointer, r io.Reader) error {
	// I really don't like doubling up on the Struct encodable, but wow,
	// what a difference it makes.
	checkPtr(ptr)
	for _, m := range e.fields {
		if err := m.enc.Decode(unsafe.Pointer(uintptr(ptr)+m.offset), r); err != nil {
			return err
		}
	}
	return nil
}
