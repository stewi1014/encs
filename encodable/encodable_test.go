package encodable_test

import (
	"bytes"
	"io"
	"reflect"
	"testing"
	"unsafe"

	"github.com/stewi1014/encs/encodable"
)

func isNil(ty reflect.Value) bool {
	switch ty.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Slice, reflect.Ptr, reflect.Map:
		return ty.IsNil()
	default:
		return false
	}
}

func testGeneric(v, want interface{}, e encodable.Encodable, t *testing.T) {
	val := reflect.ValueOf(v).Elem()
	ptr := unsafe.Pointer(val.UnsafeAddr())

	if e.Type() != val.Type() {
		t.Errorf("Type() returns %v but type to encode is %v", e.Type(), val.Type())
	}

	buff := new(bytes.Buffer)

	err := e.Encode(ptr, buff)
	if err != nil {
		t.Error(err)
	}

	if size := e.Size(); size > 0 {
		if buff.Len() > size {
			t.Errorf("Size() returns %v but %v bytes were written", size, buff.Len())
		}
	}

	decodedValue := reflect.New(val.Type()).Elem()
	decodedPtr := unsafe.Pointer(decodedValue.UnsafeAddr())
	err = e.Decode(decodedPtr, buff)
	if err != nil {
		t.Error(err)
	}

	w := reflect.ValueOf(want).Elem()

	if !reflect.DeepEqual(w.Interface(), decodedValue.Interface()) {
		wNil := isNil(w)
		dNil := isNil(decodedValue)
		t.Errorf("%v (%v, nil: %v) and %v (%v, nil: %v) are not equal", w.Type(), w, wNil, decodedValue.Type(), decodedValue, dNil)
	}

	if buff.Len() > 0 {
		t.Errorf("data remaining in buffer %v", buff.Bytes())
	}
}

type buffer struct {
	buff []byte
	off  int
}

// reset resets reading
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
