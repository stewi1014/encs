package encodable_test

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/maxatome/go-testdeep/td"
	"github.com/stewi1014/encs/encodable"
)

func TestFloat32(t *testing.T) {
	testCases := []float32{
		0, 1, 2, 3, 4, 5, 6, 254, 255, 256, 1<<32 - 1,
	}

	enc := encodable.NewFloat32(reflect.TypeOf(float32(0)))

	for _, tC := range testCases {
		t.Run(fmt.Sprint(tC), func(t *testing.T) {
			testEqual(&tC, &tC, enc, t)
		})
	}
}

func TestFloat64(t *testing.T) {
	testCases := []float64{
		0, 1, 2, 3, 4, 5, 6, 254, 255, 256, 1<<32 - 1,
	}

	enc := encodable.NewFloat64(reflect.TypeOf(float64(0)))

	for _, tC := range testCases {
		t.Run(fmt.Sprint(tC), func(t *testing.T) {
			testEqual(&tC, &tC, enc, t)
		})
	}
}

func TestComplex64(t *testing.T) {
	testCases := []complex64{
		0 + 4i, 112 + 31i, 1<<89 - 1 + 1384603i, 3, 4i,
	}

	enc := encodable.NewComplex64(reflect.TypeOf(complex64(0)))

	for _, tC := range testCases {
		t.Run(fmt.Sprint(tC), func(t *testing.T) {
			testEqual(&tC, &tC, enc, t)
		})
	}
}

func TestComplex128(t *testing.T) {
	testCases := []complex128{
		0 + 4i, 112 + 31i, 1<<89 - 1 + 1384603i, 3, 4i,
	}

	enc := encodable.NewComplex128(reflect.TypeOf(complex128(0)))

	for _, tC := range testCases {
		t.Run(fmt.Sprint(tC), func(t *testing.T) {
			testEqual(&tC, &tC, enc, t)
		})
	}
}

func TestVarComplex(t *testing.T) {
	testCases := []struct {
		desc string
		enc  interface{}
		dec  interface{}
	}{
		{
			desc: "128 to 64",
			enc:  complex128(3 + 4i),
			dec:  complex64(3 + 4i),
		},
		{
			desc: "64 to 64",
			enc:  complex64(3 + 4i),
			dec:  complex64(3 + 4i),
		},
		{
			desc: "64 to 128",
			enc:  complex64(3 + 4i),
			dec:  complex128(3 + 4i),
		},
		{
			desc: "128 to 128",
			enc:  complex128(3 + 4i),
			dec:  complex128(3 + 4i),
		},
	}

	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {

			decodeValue := reflect.New(reflect.TypeOf(tC.dec)).Elem()
			encodedValue := reflect.New(reflect.TypeOf(tC.enc)).Elem()
			encodedValue.Set(reflect.ValueOf(tC.enc))

			enc := encodable.NewVarComplex(encodedValue.Type())
			dec := encodable.NewVarComplex(decodeValue.Type())

			runTestNoErr(encodedValue, decodeValue, enc, dec, t)

			td.Cmp(t, decodeValue.Interface(), tC.dec)
		})
	}
}
