package encodable_test

import (
	"bytes"
	"fmt"
	"io"
	"reflect"
	"testing"
	"unsafe"

	"github.com/kr/pretty"
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

func runTest(v interface{}, enc encodable.Encodable, t *testing.T) interface{} {
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
		t.Fatal(err)
		return nil
	}

	if size := enc.Size(); size >= 0 && buff.Len() > size {
		t.Errorf("Size() returns %v but %v bytes were written", size, buff.Len())
	}

	decoded := reflect.New(val.Type()).Elem()
	err = enc.Decode(unsafe.Pointer(decoded.UnsafeAddr()), buff)
	if err != nil {
		t.Fatal(err)
		return nil
	}

	if buff.Len() > 0 {
		t.Errorf("data remaining in buffer %v", buff.Bytes())
	}

	return decoded.Addr().Interface()
}

func testGeneric(v, want interface{}, e encodable.Encodable, t *testing.T) {
	got := runTest(v, e, t)

	if !reflect.DeepEqual(got, want) {
		t.Errorf("%v (%v) and %v (%v) are not equal", reflect.TypeOf(want), reflect.ValueOf(want).String(), reflect.TypeOf(want), reflect.ValueOf(want).String())
		fmt.Println("Got:")
		fmt.Println(pretty.Sprint(got))
		fmt.Println("Wanted:")
		fmt.Println(pretty.Sprint(want))
	}
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
