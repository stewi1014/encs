package encodable_test

import (
	"reflect"
	"testing"

	"github.com/stewi1014/encs/encodable"
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
			enc := encodable.NewSlice(reflect.TypeOf(tC.encode).Elem(), encodable.SourceFromFunc(func(t reflect.Type, s encodable.Source) encodable.Encodable {
				switch t.Kind() {
				case reflect.Int:
					return encodable.NewInt(t)
				case reflect.String:
					return encodable.NewString(t)
				}
				return nil
			}))

			testEqual(tC.encode, tC.encode, enc, t)
		})
	}
}
