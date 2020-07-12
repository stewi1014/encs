package encodable_test

import (
	"io/ioutil"
	"reflect"
	"testing"
	"time"
	"unsafe"

	"github.com/stewi1014/encs/encodable"
)

func TestPointer(t *testing.T) {
	testCases := []struct {
		desc   string
		encode interface{}
	}{
		{
			desc: "Nil Pointer",
			encode: func() **int {
				var i *int
				return &i
			}(),
		},
		{
			desc: "Non-nil Pointer",
			encode: func() **int {
				i := new(int)
				return &i
			}(),
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			e := encodable.NewPointer(reflect.TypeOf(tC.encode).Elem(), 0, &encodable.DefaultSource{})
			testGeneric(tC.encode, tC.encode, e, t)
		})
	}
}

func TestMap(t *testing.T) {
	testCases := []struct {
		desc   string
		encode interface{}
	}{
		{
			desc: "Nil Map",
			encode: func() *map[string]int {
				var m map[string]int
				return &m
			}(),
		},
		{
			desc: "0 Length Map",
			encode: func() *map[string]int {
				m := make(map[string]int)
				return &m
			}(),
		},
		{
			desc: "Non-nil [string]int map",
			encode: func() *map[string]int {
				m := map[string]int{
					"Hello": 1,
					"World": 2,
				}
				return &m
			}(),
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			e := encodable.NewMap(reflect.TypeOf(tC.encode).Elem(), 0, &encodable.DefaultSource{})
			testGeneric(tC.encode, tC.encode, e, t)
		})
	}
}

func TestInterface(t *testing.T) {
	testCases := []struct {
		desc   string
		encode interface{}
	}{
		{
			desc: "Nil Interface",
			encode: func() *interface{} {
				var m interface{}
				return &m
			}(),
		},
		{
			desc: "Int Interface",
			encode: func() *interface{} {
				var m interface{} = int(1)
				return &m
			}(),
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			e := encodable.NewInterface(reflect.TypeOf(tC.encode).Elem(), 0, &encodable.DefaultSource{})
			testGeneric(tC.encode, tC.encode, e, t)
		})
	}
}

func TestSlice(t *testing.T) {
	testCases := []struct {
		desc   string
		encode interface{}
	}{
		{
			desc: "Nil Slice",
			encode: func() *[]int {
				var m []int
				return &m
			}(),
		},
		{
			desc: "Empty Int slice",
			encode: func() *[]int {
				m := []int{}
				return &m
			}(),
		},
		{
			desc: "A few numbers",
			encode: func() *[]int {
				m := []int{
					1, 2, 3, 4,
				}
				return &m
			}(),
		},
		{
			desc: "Some strings",
			encode: func() *[]string {
				m := []string{
					"Hello",
					"World",
				}
				return &m
			}(),
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			e := encodable.NewSlice(reflect.TypeOf(tC.encode).Elem(), 0, &encodable.DefaultSource{})
			testGeneric(tC.encode, tC.encode, e, t)
		})
	}
}

func TestArray(t *testing.T) {
	testCases := []struct {
		desc   string
		encode interface{}
	}{
		{
			desc: "Empty Int array",
			encode: func() *[0]int {
				m := [0]int{}
				return &m
			}(),
		},
		{
			desc: "A few numbers",
			encode: func() *[4]int {
				m := [4]int{
					1, 2, 3, 4,
				}
				return &m
			}(),
		},
		{
			desc: "Some strings",
			encode: func() *[2]string {
				m := [2]string{
					"Hello",
					"A longer string",
				}
				return &m
			}(),
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			e := encodable.NewArray(reflect.TypeOf(tC.encode).Elem(), 0, &encodable.DefaultSource{})
			testGeneric(tC.encode, tC.encode, e, t)
		})
	}
}

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

func TestStructStrict(t *testing.T) {
	// Strict
	for _, tC := range structTestCases {
		t.Run(tC.desc, func(t *testing.T) {
			e := encodable.New(reflect.TypeOf(tC.encode).Elem(), 0)

			testGeneric(tC.encode, tC.want, e, t)
		})
	}
}

func TestStructLoose(t *testing.T) {
	// Loose
	for _, tC := range structTestCases {
		t.Run(tC.desc, func(t *testing.T) {
			e := encodable.New(reflect.TypeOf(tC.encode).Elem(), encodable.LooseTyping)

			testGeneric(tC.encode, tC.want, e, t)
		})
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

	enc := encodable.NewStruct(reflect.TypeOf(benchStruct), 0, &encodable.DefaultSource{})
	for i := 0; i < b.N; i++ {
		err := enc.Encode(unsafe.Pointer(&benchStruct), ioutil.Discard)
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

	enc := encodable.NewStruct(reflect.TypeOf(benchStruct), 0, &encodable.DefaultSource{})
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

func BenchmarkStructLooseEncode(b *testing.B) {
	benchStruct := TestStruct2{
		Name:     "9b899bec35bc6bb8",
		BirthDay: time.Date(2019, 10, 19, 12, 28, 39, 731486213, time.UTC),
		Phone:    "2000f6a906",
		Siblings: 2,
		Spouse:   false,
		Money:    0.16683100555848812,
	}

	enc := encodable.NewStruct(reflect.TypeOf(benchStruct), encodable.LooseTyping, &encodable.DefaultSource{})
	for i := 0; i < b.N; i++ {
		err := enc.Encode(unsafe.Pointer(&benchStruct), ioutil.Discard)
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

	enc := encodable.NewStruct(reflect.TypeOf(benchStruct), encodable.LooseTyping, &encodable.DefaultSource{})
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
