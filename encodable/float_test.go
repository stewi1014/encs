package encodable_test

import (
	"fmt"
	"testing"

	"github.com/stewi1014/encs/encodable"
)

func TestFloat32(t *testing.T) {
	testCases := []float32{
		0, 1, 2, 3, 4, 5, 6, 254, 255, 256, 1<<32 - 1,
	}

	enc := encodable.NewFloat32()

	for _, tC := range testCases {
		t.Run(fmt.Sprint(tC), func(t *testing.T) {
			testGeneric(&tC, &tC, enc, t)
		})
	}
}

func TestFloat64(t *testing.T) {
	testCases := []float64{
		0, 1, 2, 3, 4, 5, 6, 254, 255, 256, 1<<32 - 1,
	}

	enc := encodable.NewFloat64()

	for _, tC := range testCases {
		t.Run(fmt.Sprint(tC), func(t *testing.T) {
			testGeneric(&tC, &tC, enc, t)
		})
	}
}

func TestComplex64(t *testing.T) {
	testCases := []complex64{
		0 + 4i, 112 + 31i, 1<<89 - 1 + 1384603i, 3, 4i,
	}

	enc := encodable.NewComplex64()

	for _, tC := range testCases {
		t.Run(fmt.Sprint(tC), func(t *testing.T) {
			testGeneric(&tC, &tC, enc, t)
		})
	}
}

func TestComplex128(t *testing.T) {
	testCases := []complex128{
		0 + 4i, 112 + 31i, 1<<89 - 1 + 1384603i, 3, 4i,
	}

	enc := encodable.NewComplex128()

	for _, tC := range testCases {
		t.Run(fmt.Sprint(tC), func(t *testing.T) {
			testGeneric(&tC, &tC, enc, t)
		})
	}
}
