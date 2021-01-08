package encodable_test

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/stewi1014/encs/encodable"
)

func TestBool(t *testing.T) {
	testCases := []bool{true, false}
	enc := encodable.NewBool(reflect.TypeOf(false))
	for _, tC := range testCases {
		t.Run(fmt.Sprint(tC), func(t *testing.T) {
			testEqual(&tC, &tC, enc, t)
		})
	}
}
