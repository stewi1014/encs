package enc

import (
	"encoding"
	"reflect"
	"time"
)

// Type constants
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

	timeTimeType    = reflect.TypeOf(time.Time{})
	interfaceType   = reflect.TypeOf(new(interface{})).Elem()
	reflectTypeType = reflect.TypeOf(new(reflect.Type)).Elem()

	invalidType = reflect.TypeOf(nil)
)

// Interface type constants
var (
	binaryMarshalerIface   = reflect.TypeOf(new(encoding.BinaryMarshaler)).Elem()
	binaryUnmarshalerIface = reflect.TypeOf(new(encoding.BinaryUnmarshaler)).Elem()
	encodableIface         = reflect.TypeOf(new(Encodable)).Elem()
)

var builtin = []interface{}{
	int(0),
	int8(0),
	int16(0),
	int32(0),
	int64(0),
	uint(0),
	uint8(0),
	uint16(0),
	uint32(0),
	uint64(0),
	uintptr(0),
	float32(0),
	float64(0),
	complex64(0),
	complex128(0),
	string(""),
	bool(true),
	time.Time{},
	time.Duration(0),
}

// Name returns the name of the type and full package import path.
func Name(t reflect.Type) string {
	if t.Kind() == reflect.Ptr {
		return "*" + Name(t.Elem())
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
