package encodable_test

import (
	"reflect"
	"testing"
)

const nilPointer = -1

func DecodeMap(m reflect.Value) error {
	v := reflect.MakeMapWithSize(rMapType, int(1))
	m.Set(v)

	nKey := reflect.New(rMapType.Key()).Elem()
	nKey.Set(reflect.ValueOf(int(1)))

	nVal := reflect.New(rMapType.Elem()).Elem()
	nVal.Field(0).Set(m)
	nVal.Field(1).Set(reflect.ValueOf(int(1)))

	v.SetMapIndex(nKey, nVal)

	return nil
}

func Decode(v reflect.Value) error {

	v.Field(1).Set(reflect.ValueOf(int(1)))

	DecodeMap(v.Field(0))
	return nil
}

var (
	rStructType = reflect.TypeOf(RecursiveTest3{})
	rMapType    = reflect.TypeOf(RecursiveTest3{}.A)
)

type RecursiveTest3 struct {
	A map[int]RecursiveTest3
	B int
}

func TestRecursiveTypes(t *testing.T) {
	encode := func() interface{} {
		s := RecursiveTest3{
			B: 1,
		}

		m := make(map[int]RecursiveTest3, 1)
		s.A = m
		m[1] = s
		return &s
	}()

	val := reflect.ValueOf(encode).Elem()

	decodedValue := reflect.New(val.Type()).Elem()
	err := Decode(decodedValue)
	if err != nil {
		t.Error(err)
	}

	w := reflect.ValueOf(encode).Elem()

	if !reflect.DeepEqual(w.Interface(), decodedValue.Interface()) {
		wNil := isNil(w)
		dNil := isNil(decodedValue)
		t.Errorf("%v (%v, nil: %v) and %v (%v, nil: %v) are not equal", w.Type(), w.String(), wNil, decodedValue.Type(), decodedValue.String(), dNil)
	}
}

func isNil(ty reflect.Value) bool {
	switch ty.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Slice, reflect.Ptr, reflect.Map:
		return ty.IsNil()
	default:
		return false
	}
}
