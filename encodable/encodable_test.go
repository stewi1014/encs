package encodable_test

import (
	"bytes"
	"fmt"
	"io"
	"reflect"
	"testing"
	"unsafe"

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

type RecursiveTest1 struct {
	N *RecursiveTest1
	B string
}

type RecursiveTest2 struct {
	S []RecursiveTest2
	B string
}

type RecursiveTest3 struct {
	M map[int]RecursiveTest3
	B string
}

func TestRecursiveTypes(t *testing.T) {
	testCases := []struct {
		desc   string
		encode interface{}
	}{
		{
			desc: "Struct with empty recursive Pointer",
			encode: &RecursiveTest1{
				N: nil,
				B: "Hello",
			},
		},
		{
			desc: "Struct value recursion",
			encode: func() *RecursiveTest1 {
				a := &RecursiveTest1{
					B: "Hello",
				}
				b := &RecursiveTest1{
					B: "World",
				}
				a.N = b
				b.N = a
				return a
			}(),
		},
		{
			desc: "Slice type recursion",
			encode: &RecursiveTest2{
				S: nil,
				B: "Hello",
			},
		},
		{
			desc: "Slice value recursion",
			encode: func() *RecursiveTest2 {
				s := make([]RecursiveTest2, 1)
				s[0].S = s
				s[0].B = "Hello"
				return &s[0]
			}(),
		},
		{
			desc: "Map type recursion",
			encode: &RecursiveTest3{
				M: nil,
				B: "Hello",
			},
		},
		{
			desc: "Map value recursion",
			encode: func() *RecursiveTest3 {
				s := RecursiveTest3{
					B: "Hello",
				}

				m := make(map[int]RecursiveTest3, 1)
				s.M = m
				m[0] = s
				return &s
			}(),
		},
	}
	for _, tC := range testCases {
		for _, config := range configPermutations {
			enc := encodable.New(reflect.TypeOf(tC.encode).Elem(), config, encodable.NewRecursiveSource(encodable.New))
			t.Run(getDescription(tC.desc, config), func(t *testing.T) {
				testGeneric(tC.encode, tC.encode, enc, t)
			})
		}
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

func testGeneric(v, want interface{}, e encodable.Encodable, t *testing.T) {
	val := reflect.ValueOf(v)
	if val.Kind() != reflect.Ptr {
		t.Errorf("Test values must be passed by pointer, got %v", val.Type())
		return
	}
	val = val.Elem()

	ptr := unsafe.Pointer(val.UnsafeAddr())

	if e.Type() != val.Type() {
		t.Errorf("Type() returns %v but type to encode is %v", e.Type(), val.Type())
		return
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
