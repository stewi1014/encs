package encode_test

import (
	"reflect"
	"testing"

	"github.com/stewi1014/encs/encodable"
	"github.com/stewi1014/encs/encode"
)

type RecursiveTest1 struct {
	N *RecursiveTest1
	B string
}

type RecursiveTest2 struct {
	S []RecursiveTest2
	B string
}

type RecursiveTest3 struct {
	M map[int]RecursiveTest3
	B string
}

type RecursiveTest4 struct {
	M reflect.Value
	B reflect.Value
}

type RecursiveTest5 struct {
	N *RecursiveTest5
	I int
	P *int
}

var recursiveTestSource = encode.NewRecursiveSource(encodable.SourceFromFunc(func(t reflect.Type, s encodable.Source) encodable.Encodable {
	if t == reflect.TypeOf(new(reflect.Value)).Elem() {
		return encode.NewValue(s)
	}
	if t == reflect.TypeOf(new(reflect.Type)).Elem() {
		return encode.NewType(true)
	}
	switch t.Kind() {
	case reflect.Int:
		return encode.NewInt(t)
	case reflect.Slice:
		return encode.NewSlice(t, s)
	case reflect.String:
		return encode.NewString(t)
	case reflect.Map:
		return encode.NewMap(t, s)
	case reflect.Ptr:
		return encode.NewPointer(t, s)
	case reflect.Struct:
		return encode.NewStructStrict(t, s)
	}
	return nil
}))

func TestRecursive(t *testing.T) {
	err := encode.Register(
		reflect.TypeOf(&RecursiveTest1{}),
		reflect.TypeOf(&RecursiveTest2{}),
		reflect.TypeOf(&RecursiveTest3{}),
		reflect.TypeOf(&RecursiveTest4{}),
		reflect.TypeOf(RecursiveTest4{}),
		reflect.TypeOf(&RecursiveTest5{}),
		reflect.TypeOf(make(map[int]reflect.Value)),
		reflect.TypeOf(new(reflect.Value)),
	)

	if err != nil {
		panic(err)
	}

	testCases := []struct {
		desc   string
		encode interface{}
	}{
		{
			desc: "Struct with empty recursive Pointer",
			encode: &RecursiveTest1{
				N: nil,
				B: "Hello",
			},
		},
		{
			desc: "Struct value recursion",
			encode: func() *RecursiveTest1 {
				a := &RecursiveTest1{
					B: "Hello",
				}
				b := &RecursiveTest1{
					B: "World",
				}
				a.N = b
				b.N = a
				return a
			}(),
		},
		{
			desc: "Slice type recursion",
			encode: &RecursiveTest2{
				S: nil,
				B: "Hello",
			},
		},
		{
			desc: "Slice value recursion",
			encode: func() *RecursiveTest2 {
				s := make([]RecursiveTest2, 1)
				s[0].S = s
				s[0].B = "Hello"
				return &s[0]
			}(),
		},
		{
			desc: "Map value recursion",
			encode: func() *RecursiveTest3 {
				s := RecursiveTest3{
					B: "Hello",
				}

				m := make(map[int]RecursiveTest3, 1)
				s.M = m
				m[0] = s
				return &s
			}(),
		},
		{
			desc: "nested pointers in linked list",
			encode: func() *RecursiveTest5 {
				s := &RecursiveTest5{
					I: 1,
				}
				s.P = &s.I

				s.N = &RecursiveTest5{
					P: &s.I,
					I: 6,
				}

				return s
			}(),
		},
		{
			desc: "Map value recurse where all values are reflect.Values",
			encode: func() *reflect.Value {
				s := RecursiveTest4{
					B: reflect.ValueOf("hello"),
				}

				m := make(map[int]reflect.Value, 1)
				s.M = reflect.ValueOf(m)
				m[0] = reflect.ValueOf(s)
				sv := reflect.ValueOf(s)
				return &sv
			}(),
		},
		{
			desc: "reflect.Value which is the value of itself",
			encode: func() *reflect.Value {
				var v reflect.Value
				v = reflect.ValueOf(&v).Elem()
				return &v
			}(),
		},
	}
	for _, tC := range testCases {
		enc := encode.NewInterface(reflect.TypeOf(&tC.encode).Elem(), recursiveTestSource)

		t.Run(tC.desc, func(t *testing.T) {
			testEqual(&tC.encode, &tC.encode, enc, t)
		})
	}
}
