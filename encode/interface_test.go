package encode_test

import (
	"reflect"
	"testing"

	"github.com/stewi1014/encs/encode"
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
			enc := encode.NewInterface(reflect.TypeOf(tC.encode).Elem(), encode.SourceFromFunc(func(t reflect.Type, s encode.Source) encode.Encodable {
				switch {
				case t.Kind() == reflect.Int:
					return encode.NewInt(t)
				case t == reflect.TypeOf(new(reflect.Type)).Elem():
					return encode.NewType(true)
				default:
					return nil
				}
			}))

			testEqual(tC.encode, tC.encode, enc, t)
		})
	}
}
