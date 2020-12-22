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
		t.Run(tC.desc, func(t *testing.T) {
			enc := encodable.NewMap(reflect.TypeOf(tC.encode).Elem(), encodable.SourceFromFunc(func(t reflect.Type, s encodable.Source) encodable.Encodable {
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
