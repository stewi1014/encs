package encodable_test

import (
	"bytes"
	"errors"
	"os"
	"reflect"
	"testing"
	"unsafe"

	"github.com/maxatome/go-testdeep/td"
	"github.com/stewi1014/encs/encodable"
	"github.com/stewi1014/encs/types"
)

// TODO: Delete this file

func reflectValueSelf() reflect.Value {
	var v reflect.Value
	v = reflect.ValueOf(v)
	return v
}

var testCases = []interface{}{
	true,
	int(123),
	[]byte("hello"),
	string("hello"),
	[32]byte{},
	map[[8]byte]string{},
	new(string),
	nil,
	reflectValueSelf(),
}

func testTypes() []reflect.Type {
	s := make([]reflect.Type, len(testCases))
	for i := range testCases {
		s[i] = reflect.TypeOf(testCases[i])
	}
	return s
}

func TestMain(m *testing.M) {
	err := types.Register(testTypes()...)
	if err != nil && !errors.Is(err, types.ErrAlreadyRegistered) {
		panic(err)
	}

	os.Exit(m.Run())
}

func getDeepEqualTester(t *testing.T, state *struct {
	rvSeen map[reflect.Value]bool
}) *td.T {
	tdt := td.NewT(t)
	if state == nil {
		state = &struct {
			rvSeen map[reflect.Value]bool
		}{
			rvSeen: make(map[reflect.Value]bool),
		}
	}

	return tdt.WithCmpHooks(func(got, expected reflect.Value) bool {
		if state.rvSeen[got] && state.rvSeen[expected] {
			return true
		}

		if !got.IsValid() || !expected.IsValid() {
			return got.IsValid() == expected.IsValid()
		}

		if !got.CanInterface() || !expected.CanInterface() {
			return got.CanInterface() == expected.CanInterface()
		}

		state.rvSeen[got] = true
		state.rvSeen[expected] = true

		return getDeepEqualTester(t, state).Cmp(got.Interface(), expected.Interface())
	})
}

func runTest(encVal, decVal reflect.Value, enc, dec encodable.Encodable, t *testing.T) (encodeErr, decodeErr error) {

	if encVal.Type() != enc.Type() {
		t.Errorf("encoder returns type %v, but type to encode is %v", enc.Type().String(), encVal.Type().String())
	}

	if decVal.Type() != dec.Type() {
		t.Errorf("decoder returns type %v, but type to decode is %v", dec.Type().String(), decVal.Type().String())
	}

	if !encVal.CanAddr() || !decVal.CanAddr() {
		t.Fatalf("values given to test with must be addressable")
	}

	buff := new(bytes.Buffer)

	err := enc.Encode(unsafe.Pointer(encVal.UnsafeAddr()), buff)
	if err != nil {
		return err, nil
	}

	if size := enc.Size(); size >= 0 && buff.Len() > size {
		t.Errorf("Size() returns %v but %v bytes were written", size, buff.Len())
	}

	err = dec.Decode(unsafe.Pointer(decVal.UnsafeAddr()), buff)
	if err != nil {
		return nil, err
	}

	if buff.Len() > 0 {
		t.Errorf("data remaining in buffer after decode: %v", buff.Bytes())
	}

	return nil, nil
}

func runSingle(val interface{}, enc encodable.Encodable, t *testing.T) (got interface{}, encode, decode error) {
	encVal := reflect.ValueOf(val)
	if encVal.Kind() != reflect.Ptr {
		panic("cannot run test with value not passed by reference")
	}
	encVal = encVal.Elem()

	decVal := reflect.New(encVal.Type()).Elem()

	encode, decode = runTest(encVal, decVal, enc, enc, t)
	got = decVal.Addr().Interface()
	return
}

func testNoErr(v interface{}, enc encodable.Encodable, t *testing.T) interface{} {
	got, eerr, derr := runSingle(v, enc, t)
	if eerr != nil {
		t.Error(eerr)
	} else if derr != nil {
		t.Error(derr)
	}

	return got
}

func testEqual(v, want interface{}, e encodable.Encodable, t *testing.T) {
	tdt := getDeepEqualTester(t, nil)

	td.CmpTrue(t, tdt.Cmp(testNoErr(v, e, t), want))
}
