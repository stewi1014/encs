package encode_test

import (
	"io/ioutil"
	"reflect"
	"testing"
	"unsafe"

	"github.com/stewi1014/encs/encode"
)

// reflectValueSelf returns a reflect.Value that is the value of itself.
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

func TestType(t *testing.T) {
	testCases := testTypes()

	for _, tC := range testCases {

		desc := "<nil>"
		if tC != nil {
			desc = tC.String()
		}

		t.Run(desc, func(t *testing.T) {
			enc := encode.NewType(true)
			testEqual(&tC, &tC, enc, t)
		})
	}
}

var typeSink reflect.Type

func BenchmarkType_Decode(b *testing.B) {
	types := testTypes()
	encodeNum := len(types) * 10
	buff := new(buffer)

	// populate buffer
	e := encode.NewType(true)
	for i := 0; i < encodeNum; i++ {
		err := e.Encode(unsafe.Pointer(&types[i%len(types)]), buff)
		if err != nil {
			b.Fatal(err)
		}
	}

	b.ResetTimer()
	var j int
	for i := 0; i < b.N; i++ {
		err := e.Decode(unsafe.Pointer(&typeSink), buff)
		if err != nil {
			b.Fatal(err)
		}
		j++
		if j >= encodeNum {
			j = 0
			buff.Reset()
		}
	}
}

func BenchmarkType_Encode(b *testing.B) {
	types := testTypes()
	e := encode.NewType(true)

	b.ResetTimer()
	var j int
	for i := 0; i < b.N; i++ {
		err := e.Encode(unsafe.Pointer(&types[i%len(types)]), ioutil.Discard)
		if err != nil {
			b.Fatal(err)
		}
		j++
		if j > len(types) {
			j = 0
		}
	}
}
