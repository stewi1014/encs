package encode_test

import (
	"reflect"
	"testing"

	"github.com/stewi1014/encs/encode"
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
		t.Run(tC.desc, func(t *testing.T) {
			enc := encode.NewArray(reflect.TypeOf(tC.encode).Elem(), encode.SourceFromFunc(func(t reflect.Type, s encode.Source) encode.Encodable {
				switch t.Kind() {
				case reflect.String:
					return encode.NewString(t)
				case reflect.Int:
					return encode.NewInt(t)
				default:
					return nil
				}
			}))

			testEqual(tC.encode, tC.encode, enc, t)
		})
	}
}
