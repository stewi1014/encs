package enc_test

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"reflect"
	"testing"
	"time"
	"unsafe"

	"github.com/stewi1014/encs/enc"
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
		config *enc.Config
		encode interface{}
		want   interface{}
	}{
		/*
			{
				desc: "TestStruct1",
				config: &enc.Config{
					TypeEncoder:       nil,
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
		*/
		{
			desc: "TestStruct2",
			config: &enc.Config{
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
			e := enc.NewEncodable(reflect.TypeOf(tC.encode), tC.config)
			buff := new(bytes.Buffer)

			fmt.Println(e.String())
			fmt.Println(unsafe.Pointer(tC.encode.(*TestStruct2)))

			err := enc.EncodeInterface(tC.encode, e, buff)
			if err != nil {
				t.Fatalf("encode error: %v", err)
			}

			checkSize(buff, e, t)

			var decoded interface{}
			err = enc.DecodeInterface(&decoded, e, buff)
			if err != nil {
				t.Fatalf("decode error: %v", err)
			}

			if !reflect.DeepEqual(tC.encode, decoded) {
				t.Fatalf("encoded %T:%v, got %T:%v", tC.encode, tC.encode, decoded, decoded)
			}

			if buff.Len() != 0 {
				t.Fatalf("data remainign in buffer %v", buff.Bytes())
			}
		})
	}
}

func BenchmarkStructEncode(b *testing.B) {
	testCases := []*TestStruct2{
		&TestStruct2{
			Name:     "9b899bec35bc6bb8",
			BirthDay: time.Date(2019, 10, 19, 12, 28, 39, 731486213, time.UTC),
			Phone:    "2000f6a906",
			Siblings: 2,
			Spouse:   false,
			Money:    0.16683100555848812,
		},
	}

	enc := enc.NewEncodable(reflect.TypeOf(TestStruct2{}), nil)

	var j int
	for i := 0; i < b.N; i++ {
		enc.Encode(unsafe.Pointer(&testCases[j]), ioutil.Discard)

		j++
		if j >= len(testCases) {
			j = 0
		}
	}
}
