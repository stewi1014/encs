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
		t.Run(tC.desc, func(t *testing.T) {
			enc := encodable.NewPointer(reflect.TypeOf(tC.encode).Elem(), encodable.SourceFromFunc(func(t reflect.Type, s encodable.Source) encodable.Encodable {
				switch t.Kind() {
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
