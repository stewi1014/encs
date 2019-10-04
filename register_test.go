package gneg

import (
	"bytes"
	"fmt"
	"net"
	"reflect"
	"testing"
)

var testTypes = []reflect.Type{
	reflect.TypeOf(int(0)),
	reflect.TypeOf(&net.UDPConn{}),
	reflect.TypeOf([]int{}),
	reflect.TypeOf(make(map[*int][]byte)),
	reflect.TypeOf(complex128(0)),
	reflect.TypeOf(new([13]*net.UDPConn)),
}

func TestRegisterResolver(t *testing.T) {
	enc := NewRegisterResolver(nil)
	dec := NewRegisterResolver(nil)

	for _, ty := range testTypes {
		enc.Register(ty)
		dec.Register(ty)
	}

	for _, ty := range testTypes {
		buff := new(bytes.Buffer)
		enc.Encode(ty, buff)

		decoded, err := dec.Decode(buff)
		if err != nil {
			t.Errorf("error decoding: %v", err)
		}
		if decoded != ty {
			t.Errorf("wrong type decoded, want %v but got %v", ty, decoded)
		}
	}
}

func TestCachingResolver(t *testing.T) {
	renc := NewRegisterResolver(nil)
	rdec := NewRegisterResolver(nil)

	for _, ty := range testTypes {
		renc.Register(ty)
		rdec.Register(ty)
	}

	enc := NewCachingResolver(renc)
	dec := NewCachingResolver(rdec)

	for _, ty := range testTypes {
		buff := new(bytes.Buffer)
		enc.Encode(ty, buff)

		decoded, err := dec.Decode(buff)
		if err != nil {
			t.Errorf("error decoding: %v", err)
		}
		if decoded != ty {
			t.Errorf("wrong type decoded, want %v but got %v", ty, decoded)
		}
	}
}

var typeSink reflect.Type

func BenchmarkRegisterResolver(b *testing.B) {
	enc := NewRegisterResolver(nil)
	dec := NewRegisterResolver(nil)
	buff := new(bytes.Buffer)

	for _, ty := range testTypes {
		enc.Register(ty)
		dec.Register(ty)
	}

	l := len(testTypes)
	var j int
	for i := 0; i < b.N; i++ {
		j++
		if j >= l {
			j = 0
		}
		err := enc.Encode(testTypes[j], buff)
		if err != nil {
			fmt.Println(err)
		}
		typeSink, _ = dec.Decode(buff)
		buff.Reset()
	}
}

func BenchmarkCachingResolver(b *testing.B) {
	renc := NewRegisterResolver(nil)
	rdec := NewRegisterResolver(nil)
	buff := new(bytes.Buffer)

	for _, ty := range testTypes {
		renc.Register(ty)
		rdec.Register(ty)
	}

	enc := NewCachingResolver(renc)
	dec := NewCachingResolver(rdec)

	l := len(testTypes)
	var j int
	for i := 0; i < b.N; i++ {
		j++
		if j >= l {
			j = 0
		}
		err := enc.Encode(testTypes[j], buff)
		if err != nil {
			fmt.Println(err)
		}
		typeSink, _ = dec.Decode(buff)
		buff.Reset()
	}
}
