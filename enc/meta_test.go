package enc_test

import (
	"bytes"
	"reflect"
	"testing"
	"time"

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
		{
			desc: "TestStruct2",
			config: &enc.Config{
				TypeEncoder:       nil,
				IncludeUnexported: false,
			},
			encode: TestStruct2{
				Name:     "John",
				BirthDay: time.Date(2019, 10, 14, 5, 50, 20, 0, time.UTC),
				Phone:    "7738234",
				Siblings: 0,
				Spouse:   true,
				Money:    -1 * (2 * 1000),
			},
			want: TestStruct2{
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
