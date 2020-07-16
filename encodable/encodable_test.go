package encodable_test

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"reflect"
	"testing"
	"unsafe"

	"github.com/maxatome/go-testdeep/td"
	"github.com/stewi1014/encs/encodable"
)

func permutate(buff *[]encodable.Config, options []encodable.Config, c encodable.Config) {
	if len(options) == 0 {
		*buff = append(*buff, c)
		return
	}

	permutate(buff, options[1:], c&^options[0])
	permutate(buff, options[1:], c|options[0])
}

var configPermutations = func() []encodable.Config {
	options := []encodable.Config{
		encodable.LooseTyping,
	}

	var buff []encodable.Config
	permutate(&buff, options, 0)
	return buff
}()

func getDescription(desc string, config encodable.Config) string {
	return fmt.Sprintf("%v %b", desc, config)
}

func runTest(v interface{}, enc encodable.Encodable, t *testing.T) (got interface{}, encode error, decode error) {
	val := reflect.ValueOf(v)
	if val.Kind() != reflect.Ptr {
		panic("cannot run test with value not passed by reference")
	}
	val = val.Elem()

	if enc.Type() != val.Type() {
		t.Errorf("Type() returns %v but type to encode is %v", enc.Type(), val.Type())
	}

	buff := new(bytes.Buffer)

	err := enc.Encode(unsafe.Pointer(val.UnsafeAddr()), buff)
	if err != nil {
		return nil, err, nil
	}

	if size := enc.Size(); size >= 0 && buff.Len() > size {
		t.Errorf("Size() returns %v but %v bytes were written", size, buff.Len())
	}

	decoded := reflect.New(val.Type()).Elem()
	err = enc.Decode(unsafe.Pointer(decoded.UnsafeAddr()), buff)
	if err != nil {
		return nil, nil, err
	}

	if buff.Len() > 0 {
		t.Errorf("data remaining in buffer %v", buff.Bytes())
	}

	return decoded.Addr().Interface(), nil, nil
}

func testEncodeError(v interface{}, want error, enc encodable.Encodable, t *testing.T) {
	_, eerr, _ := runTest(v, enc, t)
	if !errors.Is(eerr, want) {
		t.Errorf("wanted error %v on encode, but got %v instead", want, eerr)
	}
}

func testDecodeError(v interface{}, want error, enc encodable.Encodable, t *testing.T) {
	_, _, derr := runTest(v, enc, t)
	if !errors.Is(derr, want) {
		t.Errorf("wanted error %v on encode, but got %v instead", want, derr)
	}
}

func testNoErr(v interface{}, enc encodable.Encodable, t *testing.T) interface{} {
	got, eerr, derr := runTest(v, enc, t)
	if eerr != nil {
		t.Error(eerr)
	} else if derr != nil {
		t.Error(derr)
	}

	return got
}

func testEqual(v, want interface{}, e encodable.Encodable, t *testing.T) bool {
	return td.Cmp(t, testNoErr(v, e, t), want)
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
