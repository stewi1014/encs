package encodable_test

import (
	"fmt"
	"reflect"
	"testing"
	"unsafe"

	"github.com/stewi1014/encs/encio"
)

const minSingleInt = int8(-1<<7 + 9)

// Decode implements Encodable.
func DecodeInt(ptr unsafe.Pointer) error {

	*(*int)(ptr) = int(1)

	return nil
}

const nilPointer = -1

// Decode implements Encodable.
func DecodeMap(ptr unsafe.Pointer) error {
	l := 1
	m := reflect.NewAt(rMapType, ptr).Elem()

	if l == nilPointer {
		m.Set(reflect.New(rMapType).Elem())
		return nil
	}

	if uintptr(l)*(rMapType.Key().Size()+rMapType.Elem().Size()) > encio.TooBig {
		return encio.NewIOError(encio.ErrMalformed, nil, fmt.Sprintf("map size of %v is too big", l), 0)
	}

	v := reflect.MakeMapWithSize(rMapType, int(l))
	m.Set(v)

	nKey := reflect.New(rMapType.Key()).Elem()
	err := DecodeInt(unsafe.Pointer(nKey.UnsafeAddr()))
	if err != nil {
		return err
	}

	nVal := reflect.New(rMapType.Elem()).Elem()
	err = DecodeSecond(unsafe.Pointer(nVal.UnsafeAddr()))
	if err != nil {
		return err
	}

	v.SetMapIndex(nKey, nVal)

	return nil
}

// Decode implements Encodable.
func Decode(ptr unsafe.Pointer) error {
	// I really don't like doubling up on the Struct encodable, but wow,
	// what a difference it makes.

	f0Offset := rStructType.Field(0).Offset
	f1Offset := rStructType.Field(1).Offset

	mptr := unsafe.Pointer(uintptr(ptr) + f0Offset)
	seenMap = &mptr
	fmt.Println(mptr)
	if err := DecodeMap(mptr); err != nil {
		return err
	}
	if err := DecodeInt(unsafe.Pointer(uintptr(ptr) + f1Offset)); err != nil {
		return err
	}
	return nil
}

// Decode implements Encodable.
func DecodeSecond(ptr unsafe.Pointer) error {
	// I really don't like doubling up on the Struct encodable, but wow,
	// what a difference it makes.

	f0Offset := rStructType.Field(0).Offset
	f1Offset := rStructType.Field(1).Offset

	mptr := unsafe.Pointer(uintptr(ptr) + f0Offset)
	reflect.NewAt(rMapType, mptr).Elem().Set(reflect.NewAt(rMapType, *seenMap).Elem())

	if err := DecodeInt(unsafe.Pointer(uintptr(ptr) + f1Offset)); err != nil {
		return err
	}
	return nil
}

var (
	rStructType = reflect.TypeOf(RecursiveTest3{})
	rMapType    = reflect.TypeOf(RecursiveTest3{}.A)

	seenMap *unsafe.Pointer
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
	err := Decode(unsafe.Pointer(decodedValue.UnsafeAddr()))
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
