package encodable_test

import (
	"reflect"
	"testing"

	"github.com/stewi1014/encs/encodable"
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

func TestRecursive(t *testing.T) {
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
	}
	for _, config := range configPermutations {
		for _, tC := range testCases {
			src := encodable.NewRecursiveSource(encodable.DefaultSource{})
			enc := src.NewEncodable(reflect.TypeOf(tC.encode).Elem(), config, nil)
			t.Run(getDescription(tC.desc, config), func(t *testing.T) {
				testEqual(tC.encode, tC.encode, *enc, t)
			})
		}

		t.Run(getDescription("Map value recurse where all values are reflect.Values", config), func(t *testing.T) {
			s := RecursiveTest4{
				B: reflect.ValueOf("hello"),
			}

			m := make(map[int]reflect.Value, 1)
			s.M = reflect.ValueOf(m)
			m[0] = reflect.ValueOf(s)
			sv := reflect.ValueOf(s)

			src := encodable.NewRecursiveSource(encodable.DefaultSource{})
			enc := src.NewEncodable(reflect.TypeOf(sv), config, nil)
			got := testNoErr(&sv, *enc, t)

			gotval := *got.(*reflect.Value)
			gots := gotval.Interface().(RecursiveTest4)
			if gots.B.String() != "hello" {
				t.Error("struct field B is not 'hello'")
			}

			if !gots.M.IsValid() {
				t.Fatalf("map is invalid")
			}

			if gots.M.IsNil() {
				t.Fatalf("map is nil")
			}

			gotm := gots.M.Interface().(map[int]reflect.Value)
			if len(gotm) != 1 {
				t.Fatalf("map length is %v, not 1", len(gotm))
			}

			gots2 := gotm[0].Interface().(RecursiveTest4)
			gotm2 := gots2.M.Interface().(map[int]reflect.Value)

			if reflect.ValueOf(&gotm).Elem().Pointer() != reflect.ValueOf(&gotm2).Elem().Pointer() {
				t.Errorf("maps are not the same map; first is %x, and embedded is %x", reflect.ValueOf(&gotm).Elem().Pointer(), reflect.ValueOf(&gotm2).Elem().Pointer())
			}
		})

		// I'm really trying here
		t.Run(getDescription("reflect.Value which is the value of itself", config), func(t *testing.T) {
			var v reflect.Value
			v = reflect.ValueOf(&v).Elem()

			src := encodable.NewRecursiveSource(encodable.DefaultSource{})
			enc := src.NewEncodable(reflect.TypeOf(v), config, nil)

			got := testNoErr(&v, *enc, t)

			gotval := got.(*reflect.Value)
			if gotval.Type() != reflect.TypeOf(reflect.Value{}) {
				t.Error("value is of wrong type")
			}

			next := gotval.Interface().(reflect.Value)
			if gotval.UnsafeAddr() != next.UnsafeAddr() {
				t.Errorf("value is not of itself")
			}
		})
	}
}
