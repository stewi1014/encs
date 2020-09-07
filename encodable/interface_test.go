package encodable_test

import (
	"reflect"
	"testing"

	"github.com/stewi1014/encs/encodable"
)

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
		for _, config := range configPermutations {
			t.Run(getDescription(tC.desc, config), func(t *testing.T) {
				src := encodable.NewRecursiveSource(encodable.DefaultSource{})
				enc := src.NewEncodable(reflect.TypeOf(tC.encode).Elem(), config, nil)
				testEqual(tC.encode, tC.encode, *enc, t)
			})
		}
	}
}
