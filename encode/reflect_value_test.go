package encode_test

import (
	"io/ioutil"
	"reflect"
	"testing"
	"unsafe"

	"github.com/stewi1014/encs/encodable"
	"github.com/stewi1014/encs/encode"
)

func testValues() []reflect.Value {
	testCases := []interface{}{
		true,
		int(123),
		[]byte("hello"),
		string("hello"),
		[32]byte{},
		map[[8]byte]string{},
		new(string),
		nil,
		(func() reflect.Value {
			var v reflect.Value
			v = reflect.ValueOf(v)
			return v
		})(),
	}

	s := make([]reflect.Value, len(testCases))
	for i := range testCases {
		s[i] = reflect.ValueOf(testCases[i])
	}
	return s
}

var reflectValueTestSource = encodable.SourceFromFunc(func(t reflect.Type, s encodable.Source) encodable.Encodable {
	switch t.Kind() {
	case reflect.Bool:
		return encode.NewBool(t)
	case reflect.Int:
		return encode.NewInt(t)
	case reflect.Slice:
		return encode.NewSlice(t, s)
	case reflect.Uint8:
		return encode.NewUint8(t)
	case reflect.String:
		return encode.NewString(t)
	case reflect.Array:
		return encode.NewArray(t, s)
	case reflect.Map:
		return encode.NewMap(t, s)
	case reflect.Ptr:
		return encode.NewPointer(t, s)
	}
	if t == reflect.TypeOf(new(reflect.Value)).Elem() {
		return encode.NewValue(s)
	}
	if t == reflect.TypeOf(new(reflect.Type)).Elem() {
		return encode.NewType(true)
	}
	return nil
})

func TestValue(t *testing.T) {
	testCases := testValues()

	for _, tC := range testCases {
		t.Run(tC.String(), func(t *testing.T) {
			enc := encode.NewValue(reflectValueTestSource)

			testEqual(&tC, &tC, enc, t)
		})
	}
}

var valueSink reflect.Value

func BenchmarkValue_Decode(b *testing.B) {
	values := testValues()
	encodeNum := len(values) * 10
	buff := new(buffer)

	enc := encode.NewValue(reflectValueTestSource)

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
	enc := encode.NewValue(reflectValueTestSource)

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
