package encodable_test

import (
	"io/ioutil"
	"reflect"
	"testing"
	"unsafe"

	"github.com/stewi1014/encs/encodable"
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

func testValues() []reflect.Value {
	s := make([]reflect.Value, len(testCases))
	for i := range testCases {
		s[i] = reflect.ValueOf(testCases[i])
	}
	return s
}

func TestType(t *testing.T) {
	testCases := testTypes()
	reflectTypeType := reflect.TypeOf(new(reflect.Type)).Elem()

	src := encodable.NewRecursiveSource(encodable.DefaultSource{})

	for _, config := range configPermutations {
		for _, tC := range testCases {

			desc := "<nil>"
			if tC != nil {
				desc = tC.String()
			}

			t.Run(getDescription(desc, config), func(t *testing.T) {
				enc := src.NewEncodable(reflectTypeType, config, nil)
				testEqual(&tC, &tC, *enc, t)
			})
		}
	}
}

var typeSink reflect.Type

func BenchmarkType_Decode(b *testing.B) {
	types := testTypes()
	encodeNum := len(types) * 10
	buff := new(buffer)

	// populate buffer
	e := encodable.NewType(0)
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
	testCases := testValues()

	src := encodable.NewRecursiveSource(encodable.DefaultSource{})

	for _, config := range configPermutations {
		for _, tC := range testCases {
			t.Run(getDescription(tC.String(), config), func(t *testing.T) {
				enc := src.NewEncodable(reflect.TypeOf(reflect.Value{}), config, nil)
				testEqual(&tC, &tC, *enc, t)
			})
		}
	}
}

var valueSink reflect.Value

func BenchmarkValue_Decode(b *testing.B) {
	values := testValues()
	encodeNum := len(values) * 10
	buff := new(buffer)

	// populate buffer
	src := encodable.NewRecursiveSource(encodable.DefaultSource{})
	enc := src.NewEncodable(reflect.TypeOf(reflect.Value{}), 0, nil)
	for i := 0; i < encodeNum; i++ {
		err := (*enc).Encode(unsafe.Pointer(&values[i%len(values)]), buff)
		if err != nil {
			b.Fatal(err)
		}
	}

	b.ResetTimer()
	var j int
	for i := 0; i < b.N; i++ {
		err := (*enc).Decode(unsafe.Pointer(&valueSink), buff)
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

func BenchmarkValue_Encode(b *testing.B) {
	values := testValues()
	src := encodable.NewRecursiveSource(encodable.DefaultSource{})
	enc := src.NewEncodable(reflect.TypeOf(reflect.Value{}), 0, nil)

	b.ResetTimer()
	var j int
	for i := 0; i < b.N; i++ {
		err := (*enc).Encode(unsafe.Pointer(&values[i%len(values)]), ioutil.Discard)
		if err != nil {
			b.Fatal(err)
		}
		j++
		if j > len(values) {
			j = 0
		}
	}
}
