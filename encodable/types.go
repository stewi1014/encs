package encodable

import (
	"encoding"
	"reflect"
	"time"
)

// Type constants
var (
	// Builtin types
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

	// Package types
	timeTimeType     = reflect.TypeOf(time.Time{})
	timeDurationType = reflect.TypeOf(time.Duration(0))
	reflectTypeType  = reflect.TypeOf(new(reflect.Type)).Elem()
	reflectValueType = reflect.TypeOf(new(reflect.Value)).Elem()
	IDType           = reflect.TypeOf(ID{})

	// Interface type constants
	binaryMarshalerType   = reflect.TypeOf(new(encoding.BinaryMarshaler)).Elem()
	binaryUnmarshalerType = reflect.TypeOf(new(encoding.BinaryUnmarshaler)).Elem()
)
