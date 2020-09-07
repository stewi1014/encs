package encodable_test

import (
	"reflect"
	"testing"

	"github.com/stewi1014/encs/encodable"
)

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
		for _, config := range configPermutations {
			t.Run(getDescription(tC.desc, config), func(t *testing.T) {
				src := encodable.NewRecursiveSource(encodable.DefaultSource{})
				enc := src.NewEncodable(reflect.TypeOf(tC.encode).Elem(), config, nil)
				testEqual(tC.encode, tC.encode, *enc, t)
			})
		}
	}
}
