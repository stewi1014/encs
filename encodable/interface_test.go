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
		t.Run(tC.desc, func(t *testing.T) {
			enc := encodable.NewInterface(reflect.TypeOf(tC.encode).Elem(), encodable.SourceFromFunc(func(t reflect.Type, s encodable.Source) encodable.Encodable {
				switch {
				case t.Kind() == reflect.Int:
					return encodable.NewInt(t)
				case t == reflect.TypeOf(new(reflect.Type)).Elem():
					return encodable.NewType(true)
				default:
					return nil
				}
			}))

			testEqual(tC.encode, tC.encode, enc, t)
		})
	}
}
