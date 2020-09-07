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
		for _, config := range configPermutations {
			t.Run(getDescription(tC.desc, config), func(t *testing.T) {
				src := encodable.NewRecursiveSource(encodable.DefaultSource{})
				enc := src.NewEncodable(reflect.TypeOf(tC.encode).Elem(), config, nil)
				testEqual(tC.encode, tC.encode, *enc, t)
			})
		}
	}
}
