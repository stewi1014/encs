package encodable_test

import (
	"bytes"
	"io/ioutil"
	"reflect"
	"testing"
	"time"
	"unsafe"

	"github.com/stewi1014/encs/encodable"
)

type TestStruct1 struct {
	Exported1 uint
	Exported2 string
}

type TestStruct2 struct {
	Name     string
	BirthDay time.Time
	Phone    string
	Siblings int
	Spouse   bool
	Money    float64
}

func TestStruct(t *testing.T) {
	testCases := []struct {
		desc   string
		config *encodable.Config
		encode interface{}
		want   interface{}
	}{
		{
			desc: "TestStruct1",
			config: &encodable.Config{
				Resolver:          nil,
				IncludeUnexported: false,
			},
			encode: TestStruct1{
				Exported1: 6,
				Exported2: "Hello world!",
			},
			want: TestStruct1{
				Exported1: 6,
				Exported2: "Hello World!",
			},
		},
		{
			desc: "TestStruct2",
			config: &encodable.Config{
				Resolver:          nil,
				IncludeUnexported: false,
			},
			encode: &TestStruct2{
				Name:     "John",
				BirthDay: time.Date(2019, 10, 14, 5, 50, 20, 0, time.UTC),
				Phone:    "7738234",
				Siblings: 0,
				Spouse:   true,
				Money:    -1 * (2 * 1000),
			},
			want: &TestStruct2{
				Name:     "John",
				BirthDay: time.Date(2019, 10, 14, 5, 50, 20, 0, time.UTC),
				Phone:    "7738234",
				Siblings: 0,
				Spouse:   true,
				Money:    -1 * (2 * 1000),
			},
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			e := encodable.New(reflect.TypeOf(tC.encode), tC.config)
			buff := new(bytes.Buffer)

			enc := reflect.New(reflect.TypeOf(tC.encode)).Elem()
			enc.Set(reflect.ValueOf(tC.encode))

			err := e.Encode(unsafe.Pointer(enc.UnsafeAddr()), buff)
			if err != nil {
				t.Fatalf("encode error: %v", err)
			}

			checkSize(buff, e, t)

			dec := reflect.New(reflect.TypeOf(tC.encode)).Elem()

			err = e.Decode(unsafe.Pointer(dec.UnsafeAddr()), buff)
			if err != nil {
				t.Fatalf("decode error: %v", err)
			}

			decoded := dec.Interface()

			if !reflect.DeepEqual(tC.encode, decoded) {
				t.Fatalf("encoded %T:%v, got %T:%v", tC.encode, tC.encode, decoded, decoded)
			}

			if buff.Len() != 0 {
				t.Fatalf("data remaining in buffer %v", buff.Bytes())
			}
		})
	}
}

func BenchmarkStructEncode(b *testing.B) {
	benchStruct := TestStruct2{
		Name:     "9b899bec35bc6bb8",
		BirthDay: time.Date(2019, 10, 19, 12, 28, 39, 731486213, time.UTC),
		Phone:    "2000f6a906",
		Siblings: 2,
		Spouse:   false,
		Money:    0.16683100555848812,
	}

	enc := encodable.NewStruct(reflect.TypeOf(benchStruct), nil)
	for i := 0; i < b.N; i++ {
		err := enc.Encode(unsafe.Pointer(&benchStruct), ioutil.Discard)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkStructDecode(b *testing.B) {
	benchStruct := TestStruct2{
		Name:     "9b899bec35bc6bb8",
		BirthDay: time.Date(2019, 10, 19, 12, 28, 39, 731486213, time.UTC),
		Phone:    "2000f6a906",
		Siblings: 2,
		Spouse:   false,
		Money:    0.16683100555848812,
	}

	enc := encodable.NewStruct(reflect.TypeOf(benchStruct), nil)
	buff := new(buffer)
	if err := enc.Encode(unsafe.Pointer(&benchStruct), buff); err != nil {
		b.Fatal(err)
	}

	for i := 0; i < b.N; i++ {
		buff.Reset()
		err := enc.Decode(unsafe.Pointer(&benchStruct), buff)
		if err != nil {
			b.Fatal(err)
		}
	}
}
