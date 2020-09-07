package encodable_test

import (
	"reflect"
	"testing"

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
		for _, config := range configPermutations {
			t.Run(getDescription(tC.desc, config), func(t *testing.T) {
				src := encodable.NewRecursiveSource(encodable.DefaultSource{})
				enc := src.NewEncodable(reflect.TypeOf(tC.encode).Elem(), config, nil)
				testEqual(tC.encode, tC.encode, *enc, t)
			})
		}
	}
}
