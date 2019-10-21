package enc_test

import (
	"bytes"
	"io/ioutil"
	"reflect"
	"testing"

	"github.com/stewi1014/encs/enc"
)

var testValues = []interface{}{
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
	[2867]byte{},
	map[[8]byte]string{},
	new(string),
}

func testTypes() []reflect.Type {
	s := make([]reflect.Type, len(testValues))
	for i := range testValues {
		s[i] = reflect.TypeOf(testValues[i])
	}
	return s
}

func TestRegisterResolver(t *testing.T) {
	e := enc.NewRegisterResolver(nil)
	d := enc.NewRegisterResolver(nil)

	for _, ty := range testTypes() {
		e.Register(ty)
		d.Register(ty)
	}

	for _, ty := range testTypes() {
		buff := new(bytes.Buffer)
		e.Encode(ty, buff)

		decoded, err := d.Decode(ty, buff)
		if err != nil {
			t.Errorf("error decoding: %v", err)
		}
		if decoded != ty {
			t.Errorf("wrong type decoded, want %v but got %v", ty, decoded)
		}
	}
}

var typeSink reflect.Type

func BenchmarkRegisterResolverDecode(b *testing.B) {
	testTypes := testTypes()
	encodeNum := len(testTypes) * 10
	buff := new(bytes.Buffer)

	// populate buffer
	e := enc.NewRegisterResolver(nil)
	for _, tt := range testTypes {
		e.Register(tt)
	}
	for i := 0; i < encodeNum; i++ {
		e.Encode(testTypes[i%len(testTypes)], buff)
	}

	bs := buff.Bytes()

	b.ResetTimer()
	var j int
	for i := 0; i < b.N; i++ {
		typeSink, _ = e.Decode(nil, buff)
		j++
		if j >= encodeNum {
			j = 0
			buff = bytes.NewBuffer(bs)
		}
	}
}

func BenchmarkRegisterResolverEncode(b *testing.B) {
	testTypes := testTypes()
	e := enc.NewRegisterResolver(nil)

	for _, tt := range testTypes {
		e.Register(tt)
	}

	b.ResetTimer()
	var j int
	for i := 0; i < b.N; i++ {
		e.Encode(testTypes[i%len(testTypes)], ioutil.Discard)
		j++
		if j > len(testTypes) {
			j = 0
		}
	}
}
