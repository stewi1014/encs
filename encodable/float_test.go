package encodable_test

import (
	"bytes"
	"fmt"
	"testing"
	"unsafe"

	"github.com/stewi1014/encs/encodable"
)

func TestFloat32(t *testing.T) {
	testCases := []float32{
		0, 1, 2, 3, 4, 5, 6, 254, 255, 256, 1<<32 - 1,
	}

	enc := encodable.NewFloat32()
	buff := new(bytes.Buffer)

	for _, tC := range testCases {
		t.Run(fmt.Sprint(tC), func(t *testing.T) {
			err := enc.Encode(unsafe.Pointer(&tC), buff)
			if err != nil {
				t.Fatalf("Encode error: %v", err)
			}

			var d float32
			err = enc.Decode(unsafe.Pointer(&d), buff)
			if err != nil {
				t.Fatalf("Decode error: %v", err)
			}

			if d != tC {
				t.Fatalf("Encoded %v, but got %v", tC, d)
			}

			if buff.Len() != 0 {
				t.Fatalf("Data remaining in buffer: %v", buff.Bytes())
			}
		})
	}
}

func TestFloat64(t *testing.T) {
	testCases := []float64{
		0, 1, 2, 3, 4, 5, 6, 254, 255, 256, 1<<32 - 1,
	}

	enc := encodable.NewFloat64()
	buff := new(bytes.Buffer)

	for _, tC := range testCases {
		t.Run(fmt.Sprint(tC), func(t *testing.T) {
			err := enc.Encode(unsafe.Pointer(&tC), buff)
			if err != nil {
				t.Fatalf("Encode error: %v", err)
			}

			var d float64
			err = enc.Decode(unsafe.Pointer(&d), buff)
			if err != nil {
				t.Fatalf("Decode error: %v", err)
			}

			if d != tC {
				t.Fatalf("Encoded %v, but got %v", tC, d)
			}

			if buff.Len() != 0 {
				t.Fatalf("Data remaining in buffer: %v", buff.Bytes())
			}
		})
	}
}

func TestComplex64(t *testing.T) {
	testCases := []complex64{
		0 + 4i, 112 + 31i, 1<<89 - 1 + 1384603i, 3, 4i,
	}

	enc := encodable.NewComplex64()
	buff := new(bytes.Buffer)

	for _, tC := range testCases {
		t.Run(fmt.Sprint(tC), func(t *testing.T) {
			err := enc.Encode(unsafe.Pointer(&tC), buff)
			if err != nil {
				t.Fatalf("Encode error: %v", err)
			}

			var d complex64
			err = enc.Decode(unsafe.Pointer(&d), buff)
			if err != nil {
				t.Fatalf("Decode error: %v", err)
			}

			if d != tC {
				t.Fatalf("Encoded %v, but got %v", tC, d)
			}

			if buff.Len() != 0 {
				t.Fatalf("Data remaining in buffer: %v", buff.Bytes())
			}
		})
	}
}

func TestComplex128(t *testing.T) {
	testCases := []complex128{
		0 + 4i, 112 + 31i, 1<<89 - 1 + 1384603i, 3, 4i,
	}

	enc := encodable.NewComplex128()
	buff := new(bytes.Buffer)

	for _, tC := range testCases {
		t.Run(fmt.Sprint(tC), func(t *testing.T) {
			err := enc.Encode(unsafe.Pointer(&tC), buff)
			if err != nil {
				t.Fatalf("Encode error: %v", err)
			}

			var d complex128
			err = enc.Decode(unsafe.Pointer(&d), buff)
			if err != nil {
				t.Fatalf("Decode error: %v", err)
			}

			if d != tC {
				t.Fatalf("Encoded %v, but got %v", tC, d)
			}

			if buff.Len() != 0 {
				t.Fatalf("Data remaining in buffer: %v", buff.Bytes())
			}
		})
	}
}
