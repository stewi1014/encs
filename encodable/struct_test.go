package encodable_test

import (
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

type TestStruct3 struct {
	encodeme string `encs:"true"`
	NoEncode string `encs:"false"`
}

var structTestCases = []struct {
	desc   string
	encode interface{}
	want   interface{}
}{
	{
		desc: "TestStruct1",
		encode: &TestStruct1{
			Exported1: 6,
			Exported2: "Hello world!",
		},
		want: &TestStruct1{
			Exported1: 6,
			Exported2: "Hello world!",
		},
	},
	{
		desc: "TestStruct2",
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
	{
		desc: "TestStruct3",
		encode: &TestStruct3{
			encodeme: "Hello",
			NoEncode: "World!",
		},
		want: &TestStruct3{
			encodeme: "Hello",
			NoEncode: "",
		},
	},
}

func TestStruct(t *testing.T) {
	for _, tC := range structTestCases {
		for _, config := range configPermutations {
			t.Run(getDescription(tC.desc, config), func(t *testing.T) {
				src := encodable.NewRecursiveSource(encodable.DefaultSource{})
				enc := src.NewEncodable(reflect.TypeOf(tC.encode).Elem(), config, nil)
				testEqual(tC.encode, tC.want, *enc, t)
			})
		}
	}
}

func BenchmarkStructStrictEncode(b *testing.B) {
	benchStruct := TestStruct2{
		Name:     "9b899bec35bc6bb8",
		BirthDay: time.Date(2019, 10, 19, 12, 28, 39, 731486213, time.UTC),
		Phone:    "2000f6a906",
		Siblings: 2,
		Spouse:   false,
		Money:    0.16683100555848812,
	}

	src := encodable.NewRecursiveSource(encodable.DefaultSource{})
	enc := src.NewEncodable(reflect.TypeOf(benchStruct), 0, nil)
	for i := 0; i < b.N; i++ {
		err := (*enc).Encode(unsafe.Pointer(&benchStruct), ioutil.Discard)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkStructStrictDecode(b *testing.B) {
	benchStruct := TestStruct2{
		Name:     "9b899bec35bc6bb8",
		BirthDay: time.Date(2019, 10, 19, 12, 28, 39, 731486213, time.UTC),
		Phone:    "2000f6a906",
		Siblings: 2,
		Spouse:   false,
		Money:    0.16683100555848812,
	}

	src := encodable.NewRecursiveSource(encodable.DefaultSource{})
	enc := src.NewEncodable(reflect.TypeOf(benchStruct), 0, nil)
	buff := new(buffer)
	if err := (*enc).Encode(unsafe.Pointer(&benchStruct), buff); err != nil {
		b.Fatal(err)
	}

	for i := 0; i < b.N; i++ {
		buff.Reset()
		err := (*enc).Decode(unsafe.Pointer(&benchStruct), buff)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkStructLooseEncode(b *testing.B) {
	benchStruct := TestStruct2{
		Name:     "9b899bec35bc6bb8",
		BirthDay: time.Date(2019, 10, 19, 12, 28, 39, 731486213, time.UTC),
		Phone:    "2000f6a906",
		Siblings: 2,
		Spouse:   false,
		Money:    0.16683100555848812,
	}

	src := encodable.NewRecursiveSource(encodable.DefaultSource{})
	enc := src.NewEncodable(reflect.TypeOf(benchStruct), encodable.LooseTyping, nil)
	for i := 0; i < b.N; i++ {
		err := (*enc).Encode(unsafe.Pointer(&benchStruct), ioutil.Discard)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkStructLooseDecode(b *testing.B) {
	benchStruct := TestStruct2{
		Name:     "9b899bec35bc6bb8",
		BirthDay: time.Date(2019, 10, 19, 12, 28, 39, 731486213, time.UTC),
		Phone:    "2000f6a906",
		Siblings: 2,
		Spouse:   false,
		Money:    0.16683100555848812,
	}

	src := encodable.NewRecursiveSource(encodable.DefaultSource{})
	enc := src.NewEncodable(reflect.TypeOf(benchStruct), encodable.LooseTyping, nil)
	buff := new(buffer)
	if err := (*enc).Encode(unsafe.Pointer(&benchStruct), buff); err != nil {
		b.Fatal(err)
	}

	for i := 0; i < b.N; i++ {
		buff.Reset()
		err := (*enc).Decode(unsafe.Pointer(&benchStruct), buff)
		if err != nil {
			b.Fatal(err)
		}
	}
}
