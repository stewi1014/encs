package gneg

import (
	"encoding"
	"fmt"
	"reflect"
	"sort"
	"unicode"
	"unicode/utf8"
	"unsafe"

	"github.com/stewi1014/gneg/gram"
)

var (
	binaryMarshaler   = reflect.TypeOf((*encoding.BinaryMarshaler)(nil)).Elem()
	binaryUnmarshaler = reflect.TypeOf((*encoding.BinaryUnmarshaler)(nil)).Elem()
)

type etype interface {
	EncType() reflect.Type
	Decode(reflect.Value, *gram.Gram) error
	Encode(reflect.Value, *gram.Gram) error
}

type sizedetype interface {
	etype
	Size() int // Negative size means undefined size. 0 is valid
}

func newetype(t reflect.Type) (etype, error) {
	if t.Implements(binaryMarshaler) && t.Implements(binaryUnmarshaler) {
		return newMarshalingType(t), nil
	}

	switch t.Kind() {
	case reflect.Ptr:
		return newPtrType(t)
	case reflect.Struct:
		return newStructType(t)
	case reflect.Array, reflect.Slice:
		return newArrayType(t)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return newIntType(t), nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return newUintType(t), nil
	case reflect.Float32, reflect.Float64:
		return newFloatType(t), nil
	case reflect.Bool:
		return newBoolType(t), nil
	case reflect.String:
		return newStringType(t), nil
	case reflect.Map:
		return newMapType(t)
	case reflect.Complex64, reflect.Complex128:
		return newComplexType(t), nil
	}

	return nil, fmt.Errorf("cannot marshal %v", t)
}

func newBaseType(t reflect.Type) baseType {
	return baseType{
		t: t,
	}
}

type baseType struct {
	t reflect.Type
}

func (t baseType) EncType() reflect.Type {
	return t.t
}

// Pointer

func newPtrType(t reflect.Type) (etype, error) {
	elem, err := newetype(t.Elem())
	if err != nil {
		return ptrType{}, err
	}
	return ptrType{
		baseType: newBaseType(t),
		elem:     elem,
	}, nil
}

type ptrType struct {
	baseType
	elem etype
}

func (t ptrType) Size() int {
	if selem, ok := t.elem.(sizedetype); ok {
		return selem.Size()
	}
	return -1
}

func (t ptrType) Decode(v reflect.Value, g *gram.Gram) error {
	vt := v.Type()
	if vt != t.t {
		return fmt.Errorf("wrong pointer type, want %v but got %v", t.t, vt)
	}
	if g.Len() == 0 {
		v.Set(reflect.New(t.t).Elem())
		return nil
	}
	if v.IsNil() {
		v.Set(reflect.New(t.elem.EncType()))
	}
	return t.elem.Decode(v.Elem(), g)
}

func (t ptrType) Encode(v reflect.Value, g *gram.Gram) error {
	if v.IsNil() {
		return nil
	}
	return t.elem.Encode(v.Elem(), g)
}

// BinaryMarshaller

func newMarshalingType(t reflect.Type) marshalingType {
	return marshalingType{
		baseType: newBaseType(t),
	}
}

type marshalingType struct {
	baseType
}

func (t marshalingType) Decode(v reflect.Value, g *gram.Gram) error {
	var i interface{} = v.Interface()
	if v.Kind() == reflect.Ptr && v.IsNil() {
		i = reflect.New(v.Type().Elem()).Interface()
	}
	if dec, ok := i.(encoding.BinaryUnmarshaler); ok {
		err := dec.UnmarshalBinary(g.ReadAll())
		v.Set(reflect.ValueOf(i))
		return err
	}
	return fmt.Errorf("%v:%v doesn't support binary unmarshalling", v.Type(), v)
}

func (t marshalingType) Encode(v reflect.Value, g *gram.Gram) error {
	i := v.Interface()
	if enc, ok := i.(encoding.BinaryMarshaler); ok {
		buff, err := enc.MarshalBinary()
		if err != nil {
			return err
		}
		_, err = g.Write(buff)
		return err
	}
	return fmt.Errorf("%v:%v doesn't support binary marshaling", v.Type(), v)
}

// Map

func newMapType(t reflect.Type) (mapType, error) {
	keyet, err := newetype(t.Key())
	if err != nil {
		return mapType{}, err
	}

	valet, err := newetype(t.Elem())
	if err != nil {
		return mapType{}, err
	}

	return mapType{
		baseType: newBaseType(t),
		key:      keyet,
		val:      valet,
	}, nil
}

type mapType struct {
	baseType
	key etype
	val etype
}

func (m mapType) Encode(v reflect.Value, g *gram.Gram) error {
	mit := v.MapRange()
	for mit.Next() {
		kh := gram.WriteSizeHeader(g)
		err := m.key.Encode(mit.Key(), g)
		if err != nil {
			return err
		}
		kh()
		vh := gram.WriteSizeHeader(g)
		err = m.val.Encode(mit.Value(), g)
		if err != nil {
			return err
		}
		vh()
	}
	return nil
}

func (m mapType) Decode(v reflect.Value, g *gram.Gram) error {
	nm := reflect.MakeMap(m.t)
	for g.Len() > 0 {
		key := reflect.New(m.key.EncType()).Elem()
		err := m.key.Decode(key, gram.ReadSizeHeader(g))
		if err != nil {
			return err
		}

		val := reflect.New(m.val.EncType()).Elem()
		err = m.val.Decode(val, gram.ReadSizeHeader(g))
		if err != nil {
			return err
		}
		nm.SetMapIndex(key, val)
	}
	v.Set(nm)
	return nil
}

// Struct

func newStructType(t reflect.Type) (structType, error) {
	st := structType{
		baseType: newBaseType(t),
	}
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)

		r, _ := utf8.DecodeRuneInString(f.Name)
		if unicode.IsLower(r) {
			continue
		}

		et, err := newetype(f.Type)
		if err != nil {
			return st, err
		}

		st.members = append(st.members, structMember{
			Name:  f.Name,
			etype: et,
		})
	}

	sort.Sort(st.members)

	return st, nil
}

type structType struct {
	baseType
	members structMembers //must be alphabetical
}

type structMember struct {
	etype
	Name string
}
type structMembers []structMember

func (a structMembers) Len() int           { return len(a) }
func (a structMembers) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a structMembers) Less(i, j int) bool { return a[i].Name < a[j].Name }

func (s structType) Encode(v reflect.Value, g *gram.Gram) error {
	for _, sm := range s.members {
		header := gram.WriteSizeHeader(g)
		err := sm.Encode(v.FieldByName(sm.Name), g)
		if err != nil {
			return err
		}
		header()
	}
	return nil
}

func (s structType) Decode(v reflect.Value, g *gram.Gram) error {
	for _, sm := range s.members {
		err := sm.Decode(v.FieldByName(sm.Name), gram.ReadSizeHeader(g))
		if err != nil {
			return err
		}
	}
	return nil
}

// Array type
// two encoders for array types. one for arrays with elements of a constant size,
// and one for arrays with elements of variable size.
func newArrayType(t reflect.Type) (etype, error) {
	elem, err := newetype(t.Elem())
	if err != nil {
		return nil, err
	}

	if selem, ok := elem.(sizedetype); ok && selem.Size() >= 0 {
		return constArrayType{
			baseType: newBaseType(t),
			elem:     selem,
		}, nil
	}

	return varArrayType{
		baseType: newBaseType(t),
		elem:     elem,
	}, nil
}

type constArrayType struct {
	baseType
	elem sizedetype
}

func (t constArrayType) Encode(v reflect.Value, g *gram.Gram) error {
	if v.Kind() == reflect.Slice && v.IsNil() {
		g.WriteUint(0)
		return nil
	}
	g.WriteUint(uint64(v.Len()))
	size := t.elem.Size()
	for i := 0; i < v.Len(); i++ {
		err := t.elem.Encode(v.Index(i), g.WriteLater(size))
		if err != nil {
			return err
		}
	}
	return nil
}

func (t constArrayType) Decode(v reflect.Value, g *gram.Gram) error {
	l := int(g.ReadUint())
	if l == 0 {
		v.Set(reflect.New(t.t).Elem())
	}

	//make sure we've got a good slice
	if (v.Kind() == reflect.Slice && v.IsNil()) || v.Cap() < l {
		// we must allocate
		if t.t.Kind() == reflect.Slice {
			v.Set(reflect.MakeSlice(t.t, l, l))
		} else {
			v.Set(reflect.New(t.t).Elem())
		}
	} else if v.Len() != l {
		v.SetLen(l)
	}

	for i := 0; i < l; i++ {
		err := t.elem.Decode(v.Index(i), g.LimitReader(t.elem.Size()))
		if err != nil {
			return err
		}
	}
	return nil
}

func (t constArrayType) Size() int {
	if t.t.Kind() == reflect.Array {
		return (t.elem.Size() * t.t.Len()) + 6
	}
	return -1
}

type varArrayType struct {
	baseType
	elem etype
}

func (t varArrayType) Encode(v reflect.Value, g *gram.Gram) error {
	panic("ni")
}

func (t varArrayType) Decode(v reflect.Value, g *gram.Gram) error {
	panic("ni")
}

// IntType

func newIntType(t reflect.Type) intType {
	return intType{
		baseType: newBaseType(t),
	}
}

type intType struct {
	baseType
}

func (t intType) Encode(v reflect.Value, g *gram.Gram) error {
	switch t.t.Kind() {
	case reflect.Int8:
		g.WriteBuff(1)[0] = uint8(v.Int())
	case reflect.Int16:
		binEnc.PutUint16(g.WriteBuff(2), uint16(v.Int()))
	case reflect.Int32:
		binEnc.PutUint32(g.WriteBuff(4), uint32(v.Int()))
	case reflect.Int, reflect.Int64:
		binEnc.PutUint64(g.WriteBuff(8), uint64(v.Int()))
	}
	return nil
}

func (t intType) Decode(v reflect.Value, g *gram.Gram) error {
	var set int64
	switch t.t.Kind() {
	case reflect.Int8:
		set = int64(int8(g.ReadBuff(1)[0]))
	case reflect.Int16:
		set = int64(int16(binEnc.Uint16(g.ReadBuff(2))))
	case reflect.Int32:
		set = int64(int32(binEnc.Uint32(g.ReadBuff(4))))
	case reflect.Int, reflect.Int64:
		set = int64(binEnc.Uint64(g.ReadBuff(8)))
	}
	v.SetInt(set)
	return nil
}

func (t intType) Size() int {
	switch t.t.Kind() {
	case reflect.Int8:
		return 1
	case reflect.Int16:
		return 2
	case reflect.Int32:
		return 4
	case reflect.Int, reflect.Int64:
		return 8
	}
	return -1
}

// UintType

func newUintType(t reflect.Type) uintType {
	return uintType{
		baseType: newBaseType(t),
	}
}

type uintType struct {
	baseType
}

func (t uintType) Encode(v reflect.Value, g *gram.Gram) error {
	switch t.t.Kind() {
	case reflect.Uint8:
		g.WriteBuff(1)[0] = uint8(v.Uint())
	case reflect.Uint16:
		binEnc.PutUint16(g.WriteBuff(2), uint16(v.Uint()))
	case reflect.Uint32:
		binEnc.PutUint32(g.WriteBuff(4), uint32(v.Uint()))
	case reflect.Uint, reflect.Uint64:
		binEnc.PutUint64(g.WriteBuff(8), uint64(v.Uint()))
	}
	return nil
}

func (t uintType) Decode(v reflect.Value, g *gram.Gram) error {
	var set uint64
	switch t.t.Kind() {
	case reflect.Uint8:
		set = uint64(g.ReadBuff(1)[0])
	case reflect.Uint16:
		set = uint64(binEnc.Uint16(g.ReadBuff(2)))
	case reflect.Uint32:
		set = uint64(binEnc.Uint32(g.ReadBuff(4)))
	case reflect.Uint, reflect.Uint64:
		set = uint64(binEnc.Uint64(g.ReadBuff(8)))
	}
	v.SetUint(set)
	return nil
}

func (t uintType) Size() int {
	switch t.t.Kind() {
	case reflect.Uint8:
		return 1
	case reflect.Uint16:
		return 2
	case reflect.Uint32:
		return 4
	case reflect.Uint, reflect.Uint64:
		return 8
	}
	return -1
}

// Bool Type

func newBoolType(t reflect.Type) boolType {
	return boolType{
		baseType: newBaseType(t),
	}
}

type boolType struct {
	baseType
}

func (t boolType) Encode(v reflect.Value, g *gram.Gram) error {
	if v.Bool() {
		g.WriteBuff(1)[0] = 1
	} else {
		g.WriteBuff(1)[0] = 0
	}
	return nil
}

func (t boolType) Decode(v reflect.Value, g *gram.Gram) error {
	n := g.ReadBuff(1)[0]
	if n > 0 {
		v.SetBool(true)
	} else {
		v.SetBool(false)
	}
	return nil
}

func (t boolType) Size() int {
	return 1
}

// String Type

func newStringType(t reflect.Type) stringType {
	return stringType{
		baseType: newBaseType(t),
	}
}

type stringType struct {
	baseType
}

func (t stringType) Encode(v reflect.Value, g *gram.Gram) error {
	g.Write(([]byte)(v.String()))
	return nil
}

func (t stringType) Decode(v reflect.Value, g *gram.Gram) error {
	v.SetString((string)(g.ReadAll()))
	return nil
}

// Float Type

func newFloatType(t reflect.Type) floatType {
	return floatType{
		baseType: newBaseType(t),
	}
}

type floatType struct {
	baseType
}

func (t floatType) Encode(v reflect.Value, g *gram.Gram) error {
	binEnc.PutUint64(g.WriteBuff(8), fbits(v.Float()))
	return nil
}

func (t floatType) Decode(v reflect.Value, g *gram.Gram) error {
	v.SetFloat(bitsf(binEnc.Uint64(g.ReadBuff(8))))
	return nil
}

func (t floatType) Size() int {
	return 8
}

func fbits(f float64) uint64 {
	return *(*uint64)(unsafe.Pointer(&f))
}

func bitsf(n uint64) float64 {
	return *(*float64)(unsafe.Pointer(&n))
}

// Complex Type

func newComplexType(t reflect.Type) complexType {
	return complexType{
		baseType: newBaseType(t),
	}
}

type complexType struct {
	baseType
}

func (t complexType) Encode(v reflect.Value, g *gram.Gram) error {
	c := v.Complex()
	binEnc.PutUint64(g.WriteBuff(8), fbits(real(c)))
	binEnc.PutUint64(g.WriteBuff(8), fbits(imag(c)))
	return nil
}

func (t complexType) Decode(v reflect.Value, g *gram.Gram) error {
	r := bitsf(binEnc.Uint64(g.ReadBuff(8)))
	i := bitsf(binEnc.Uint64(g.ReadBuff(8)))
	v.SetComplex(complex(r, i))
	return nil
}

func (t complexType) Size() int {
	if t.t.Kind() == reflect.Complex128 {
		return 16
	}
	return 8
}
