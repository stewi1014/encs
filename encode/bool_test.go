package encode_test

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/stewi1014/encs/encode"
)

func TestBool(t *testing.T) {
	testCases := []bool{true, false}
	enc := encode.NewBool(reflect.TypeOf(false))
	for _, tC := range testCases {
		t.Run(fmt.Sprint(tC), func(t *testing.T) {
			testEqual(&tC, &tC, enc, t)
		})
	}
}
