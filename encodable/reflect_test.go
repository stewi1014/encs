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

var reflectTestSource = encodable.SourceFromFunc(func(t reflect.Type, s encodable.Source) encodable.Encodable {
	switch t.Kind() {
	case reflect.Bool:
		return encodable.NewBool(t)
	case reflect.Int:
		return encodable.NewInt(t)
	case reflect.Slice:
		return encodable.NewSlice(t, s)
	case reflect.Uint8:
		return encodable.NewUint8(t)
	case reflect.String:
		return encodable.NewString(t)
	case reflect.Array:
		return encodable.NewArray(t, s)
	case reflect.Map:
		return encodable.NewMap(t, s)
	case reflect.Ptr:
		return encodable.NewPointer(t, s)
	}
	if t == reflect.TypeOf(new(reflect.Value)).Elem() {
		return encodable.NewValue(s)
	}
	if t == reflect.TypeOf(new(reflect.Type)).Elem() {
		return encodable.NewType(true)
	}
	return nil
})

func TestType(t *testing.T) {
	testCases := testTypes()

	for _, tC := range testCases {

		desc := "<nil>"
		if tC != nil {
			desc = tC.String()
		}

		t.Run(desc, func(t *testing.T) {
			enc := encodable.NewType(true)
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
	e := encodable.NewType(true)
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
	e := encodable.NewType(true)

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

	for _, tC := range testCases {
		t.Run(tC.String(), func(t *testing.T) {
			enc := encodable.NewValue(reflectTestSource)

			testEqual(&tC, &tC, enc, t)
		})
	}
}

var valueSink reflect.Value

func BenchmarkValue_Decode(b *testing.B) {
	values := testValues()
	encodeNum := len(values) * 10
	buff := new(buffer)

	enc := encodable.NewValue(reflectTestSource)

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
	enc := encodable.NewValue(reflectTestSource)

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
