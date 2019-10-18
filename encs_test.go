package encs_test

import (
	"bytes"
	"encoding/gob"
	"reflect"
	"testing"

	"github.com/stewi1014/encs"
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
}

func testTypes() []reflect.Type {
	s := make([]reflect.Type, len(testValues))
	for i := range testValues {
		s[i] = reflect.TypeOf(testValues[i])
	}
	return s
}

func TestBasicEncodeDecode(t *testing.T) {
	buff := new(bytes.Buffer)

	enc := encs.NewEncoder(buff)
	dec := encs.NewDecoder(buff)

	var decoded interface{}
	for _, v := range testValues {
		err := enc.Encode(v)
		if err != nil {
			t.Errorf("encode error: %v", err)
			return
		}

		err = dec.DecodeInterface(&decoded)
		if err != nil {
			t.Errorf("decode error: %v", err)
			return
		}

		if buff.Len() != 0 {
			t.Errorf("data remaining in buffer %v", buff.Bytes())
		}

		if !reflect.DeepEqual(v, decoded) {
			t.Errorf("encoding %T:%v, decoding got %v", v, v, decoded)
		}
	}
}

func BenchmarkStream(b *testing.B) {
	buff := new(bytes.Buffer)
	enc := encs.NewEncoder(buff)
	dec := encs.NewDecoder(buff)

	var v interface{}
	j := 0
	for i := 0; i < b.N; i++ {
		err := enc.Encode(testValues[j])
		if err != nil {
			b.Fatal(err)
		}
		err = dec.DecodeInterface(&v)
		if err != nil {
			b.Fatal(err)
		}

		j++
		if j == len(testValues) {
			j = 0
		}
	}
}

func BenchmarkGob(b *testing.B) {
	buff := new(bytes.Buffer)
	enc := gob.NewEncoder(buff)
	dec := gob.NewDecoder(buff)

	var v interface{}
	j := 0
	for i := 0; i < b.N; i++ {
		enc.Encode(testValues[j])
		dec.Decode(&v)

		j++
		if j == len(testValues) {
			j = 0
		}
	}
}
