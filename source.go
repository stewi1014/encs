package encs

import (
	"encoding"
	"fmt"
	"reflect"

	"github.com/stewi1014/encs/encio"
	"github.com/stewi1014/encs/encodable"
)

var (
	reflectTypeType       = reflect.TypeOf(new(reflect.Type)).Elem()
	reflectValueType      = reflect.TypeOf(new(reflect.Value)).Elem()
	binaryMarshalerType   = reflect.TypeOf(new(encoding.BinaryMarshaler)).Elem()
	binaryUnmarshalerType = reflect.TypeOf(new(encoding.BinaryUnmarshaler)).Elem()
)

// DefaultSource is a simple Source for Encodables. It performs no pointer logic.
// Use RecursiveSource unless it is guaranteed there will not be recursive types or values encoded,
// and the pointer reference structure doesn't matter. i.e. If a struct Encodable is created with an int and *int field
// where the *int field points to the int field, the decoded *int field will not point to the struct's own field.
// It is also slower for large types.
// DefaultSource{} is an appropriate way to instantiate it.
var DefaultSource = encodable.SourceFromFunc(func(t reflect.Type, s encodable.Source) encodable.Encodable {

	ptrt := reflect.PtrTo(t)
	kind := t.Kind()
	switch {
	// Implementers
	case ptrt.Implements(binaryMarshalerType) && ptrt.Implements(binaryUnmarshalerType):
		return encodable.NewBinaryMarshaler(t)

	// Specific types
	case t == reflectTypeType:
		return encodable.NewType(false)
	case t == reflectValueType:
		return encodable.NewValue(s)

	// Compound-Types
	case kind == reflect.Ptr:
		return encodable.NewPointer(t, s)
	case kind == reflect.Interface:
		return encodable.NewInterface(t, s)
	case kind == reflect.Struct:
		return encodable.NewStructLoose(t, s)
	case kind == reflect.Array:
		return encodable.NewArray(t, s)
	case kind == reflect.Slice:
		return encodable.NewSlice(t, s)
	case kind == reflect.Map:
		return encodable.NewMap(t, s)

	// Number types
	case kind == reflect.Uint8,
		kind == reflect.Uint16,
		kind == reflect.Uint32,
		kind == reflect.Uint64,
		kind == reflect.Uint,
		kind == reflect.Int8,
		kind == reflect.Int16,
		kind == reflect.Int32,
		kind == reflect.Int64,
		kind == reflect.Int,
		kind == reflect.Uintptr,
		kind == reflect.Float32,
		kind == reflect.Float64:
		return encodable.NewVarint(t)

	case kind == reflect.Complex64,
		kind == reflect.Complex128:
		return encodable.NewVarComplex(t)

	// Misc types
	case kind == reflect.Bool:
		return encodable.NewBool(t)
	case kind == reflect.String:
		return encodable.NewString(t)
	default:
		panic(encio.NewError(encio.ErrBadType, fmt.Sprintf("cannot create encodable for type %v", t), 0))
	}
})
