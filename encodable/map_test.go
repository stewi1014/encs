package encodable_test

import (
	"reflect"
	"testing"

	"github.com/stewi1014/encs/encodable"
)

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
		for _, config := range configPermutations {
			t.Run(getDescription(tC.desc, config), func(t *testing.T) {
				src := encodable.NewRecursiveSource(encodable.DefaultSource{})
				enc := src.NewEncodable(reflect.TypeOf(tC.encode).Elem(), config, nil)
				testEqual(tC.encode, tC.encode, *enc, t)
			})
		}
	}
}
