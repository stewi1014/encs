package encodable_test

import (
	"bytes"
	"errors"
	"io/ioutil"
	"os"
	"reflect"
	"testing"
	"unsafe"

	"github.com/stewi1014/encs/encodable"
)

var testCases = []interface{}{
	true,
	int(123),
	int8(123),
	int16(-12345),
	int32(123456),
	int64(-1234567),
	uint(123),
	uint8(123),
	uint16(12345),
	uint32(123456),
	uint64(1234567),
	uintptr(12345678),
	float32(1.2345),
	float64(1.2345678),
	complex64(1.2345 + 2.3456i),
	complex128(1.2345678 + 2.3456789i),
	[]byte("hello"),
	string("hello"),
	[32]byte{},
	map[[8]byte]string{},
	new(string),
}

func TestMain(m *testing.M) {
	err := encodable.Register(testTypes()...)
	if err != nil && !errors.Is(err, encodable.ErrAlreadyRegistered) {
		panic(err)
	}
	os.Exit(m.Run())
}

func testTypes() []reflect.Type {
	s := make([]reflect.Type, len(testCases))
	for i := range testCases {
		s[i] = reflect.TypeOf(testCases[i])
	}
	return s
}

func testValues() []reflect.Value {
	s := make([]reflect.Value, len(testCases))
	for i := range testCases {
		s[i] = reflect.ValueOf(testCases[i])
	}
	return s
}

func TestType(t *testing.T) {
	types := testTypes()

	e := encodable.NewType(0)

	for i := range types {
		buff := new(bytes.Buffer)
		err := e.Encode(unsafe.Pointer(&types[i]), buff)
		if err != nil {
			t.Fatal(err)
		}

		var decoded reflect.Type
		err = e.Decode(unsafe.Pointer(&decoded), buff)
		if err != nil {
			t.Errorf("error decoding: %v", err)
		}
		if decoded != types[i] {
			t.Errorf("wrong type decoded, want %v but got %v", types[i], decoded)
		}
	}
}

var typeSink reflect.Type

func BenchmarkType_Decode(b *testing.B) {
	types := testTypes()
	encodeNum := len(types) * 10
	buff := new(bytes.Buffer)

	// populate buffer
	e := encodable.NewType(0)
	for i := 0; i < encodeNum; i++ {
		err := e.Encode(unsafe.Pointer(&types[i%len(types)]), buff)
		if err != nil {
			b.Fatal(err)
		}
	}

	bs := buff.Bytes()

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
			buff = bytes.NewBuffer(bs)
		}
	}
}

func BenchmarkType_Encode(b *testing.B) {
	types := testTypes()
	e := encodable.NewType(0)

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

func TestValue(t *testing.T) {
	values := testValues()

	e := encodable.NewValue(0, &encodable.DefaultSource{})

	for i := range values {
		buff := new(bytes.Buffer)
		err := e.Encode(unsafe.Pointer(&values[i]), buff)
		if err != nil {
			t.Fatal(err)
		}

		var decoded reflect.Value
		err = e.Decode(unsafe.Pointer(&decoded), buff)
		if err != nil {
			t.Errorf("error decoding: %v", err)
		}
		if reflect.DeepEqual(reflect.TypeOf(values[i]), decoded) {
			t.Errorf("wrong type decoded, want %v but got %v", values[i], decoded)
		}
	}
}

var valueSink reflect.Value

func BenchmarkValue_Decode(b *testing.B) {
	values := testValues()
	encodeNum := len(values) * 10
	buff := new(bytes.Buffer)

	// populate buffer
	e := encodable.NewValue(0, &encodable.DefaultSource{})
	for i := 0; i < encodeNum; i++ {
		err := e.Encode(unsafe.Pointer(&values[i%len(values)]), buff)
		if err != nil {
			b.Fatal(err)
		}
	}

	bs := buff.Bytes()

	b.ResetTimer()
	var j int
	for i := 0; i < b.N; i++ {
		err := e.Decode(unsafe.Pointer(&valueSink), buff)
		if err != nil {
			b.Fatal(err)
		}
		j++
		if j >= encodeNum {
			j = 0
			buff = bytes.NewBuffer(bs)
		}
	}
}

func BenchmarkValue_Encode(b *testing.B) {
	values := testValues()
	e := encodable.NewValue(0, &encodable.DefaultSource{})

	b.ResetTimer()
	var j int
	for i := 0; i < b.N; i++ {
		err := e.Encode(unsafe.Pointer(&values[i%len(values)]), ioutil.Discard)
		if err != nil {
			b.Fatal(err)
		}
		j++
		if j > len(values) {
			j = 0
		}
	}
}