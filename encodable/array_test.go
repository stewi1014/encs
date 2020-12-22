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
		t.Run(tC.desc, func(t *testing.T) {
			enc := encodable.NewArray(reflect.TypeOf(tC.encode).Elem(), encodable.SourceFromFunc(func(t reflect.Type, s encodable.Source) encodable.Encodable {
				switch t.Kind() {
				case reflect.String:
					return encodable.NewString(t)
				case reflect.Int:
					return encodable.NewInt(t)
				default:
					return nil
				}
			}))

			testEqual(tC.encode, tC.encode, enc, t)
		})
	}
}
