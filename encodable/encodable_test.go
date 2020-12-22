package encodable_test

import (
	"bytes"
	"errors"
	"io"
	"os"
	"reflect"
	"testing"
	"unsafe"

	"github.com/maxatome/go-testdeep/td"
	"github.com/stewi1014/encs/encodable"
)

func TestMain(m *testing.M) {
	err := encodable.Register(testTypes()...)
	if err != nil && !errors.Is(err, encodable.ErrAlreadyRegistered) {
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

func runTestNoErr(encVal, decVal reflect.Value, enc, dec encodable.Encodable, t *testing.T) {
	encErr, decErr := runTest(encVal, decVal, enc, dec, t)
	if encErr != nil {
		t.Error(encErr)
		return
	}
	if decErr != nil {
		t.Error(decErr)
	}
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

type buffer struct {
	buff []byte
	off  int
}

// Reset resets reading, allowing the same buffer to be read again.
func (b *buffer) Reset() {
	b.off = 0
}

func (b *buffer) Read(buff []byte) (int, error) {
	n := copy(buff, b.buff[b.off:])
	b.off += n
	if n < len(buff) {
		return n, io.EOF
	}
	return n, nil
}

func (b *buffer) Write(buff []byte) (int, error) {
	copy(b.buff[b.grow(len(buff)):], buff)
	return len(buff), nil
}

func (b *buffer) grow(n int) int {
	l := len(b.buff)
	c := cap(b.buff)
	if l+n <= c {
		b.buff = b.buff[:l+n]
		return l
	}

	nb := make([]byte, l+n, c*2+n)
	copy(nb, b.buff)
	b.buff = nb
	return l
}
