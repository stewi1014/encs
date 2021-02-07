package types

import (
	"encoding"
	"reflect"
)

var (
	ReflectTypeType  = reflect.TypeOf(new(reflect.Type)).Elem()
	ReflectValueType = reflect.TypeOf(new(reflect.Value)).Elem()

	BinaryMarshalerType   = reflect.TypeOf(new(encoding.BinaryMarshaler)).Elem()
	BinaryUnmarshalerType = reflect.TypeOf(new(encoding.BinaryUnmarshaler)).Elem()
)
