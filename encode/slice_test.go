package encode_test

import (
	"reflect"
	"testing"

	"github.com/stewi1014/encs/encode"
)

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
			enc := encode.NewSlice(reflect.TypeOf(tC.encode).Elem(), encode.SourceFromFunc(func(t reflect.Type, s encode.Source) encode.Encodable {
				switch t.Kind() {
				case reflect.Int:
					return encode.NewInt(t)
				case reflect.String:
					return encode.NewString(t)
				}
				return nil
			}))

			testEqual(tC.encode, tC.encode, enc, t)
		})
	}
}
