package encodable_test

import (
	"bytes"
	"fmt"
	"math/bits"
	"reflect"
	"testing"
	"unsafe"

	"github.com/stewi1014/encs/encodable"
)

func TestInt8(t *testing.T) {
	enc := encodable.NewInt8(reflect.TypeOf(int8(0)))
	for tC := int8(-1 << 7); ; tC++ {
		testEqual(&tC, &tC, enc, t)

		if tC == 1<<7-1 {
			break
		}
	}
}

func TestInt16(t *testing.T) {
	enc := encodable.NewInt16(reflect.TypeOf(int16(0)))
	for tC := int16(-1 << 15); ; tC++ {
		testEqual(&tC, &tC, enc, t)

		if tC == 1<<15-1 {
			break
		}
	}
}

func TestInt32(t *testing.T) {
	testCases := []int32{
		0, 1, 2, 3, 4, 5, 6, 254, 255, 256, -1 << 31, 1<<31 - 1, -1,
	}

	enc := encodable.NewInt32(reflect.TypeOf(int32(0)))

	for _, tC := range testCases {
		t.Run(fmt.Sprint(tC), func(t *testing.T) {
			testEqual(&tC, &tC, enc, t)
		})
	}
}

func TestInt64(t *testing.T) {
	testCases := []int64{
		0, 1, 2, 3, 4, 5, 6, 254, 255, 256, 1<<32 - 1, 1<<63 - 1, -1, -1 << 63,
	}

	enc := encodable.NewInt64(reflect.TypeOf(int64(0)))

	for _, tC := range testCases {
		t.Run(fmt.Sprint(tC), func(t *testing.T) {
			testEqual(&tC, &tC, enc, t)
		})
	}
}

func TestInt(t *testing.T) {
	testCases := []int{
		0, 1, 2, 3, 4, 5, 6, 254, 255, 256, 1<<32 - 1, 1<<(bits.UintSize-1) - 1, -1, -1 << 63,
	}

	enc := encodable.NewInt(reflect.TypeOf(int(0)))

	for _, tC := range testCases {
		t.Run(fmt.Sprint(tC), func(t *testing.T) {
			testEqual(&tC, &tC, enc, t)
		})
	}
}

func BenchmarkInt(b *testing.B) {
	ints := []int{
		0, 1, 2, 3, 4, 5, 6, 254, 255, 256, 1<<32 - 1, 1<<(bits.UintSize-1) - 1, -1, -1 << 63,
	}

	enc := encodable.NewInt(reflect.TypeOf(int(0)))
	buff := new(bytes.Buffer)
	j := 0
	var u int

	for i := 0; i < b.N; i++ {
		err := enc.Encode(unsafe.Pointer(&ints[j]), buff)
		if err != nil {
			b.Fatal(err)
		}
		err = enc.Decode(unsafe.Pointer(unsafe.Pointer(&u)), buff)
		if err != nil {
			b.Fatal(err)
		}

		j++
		if j >= len(ints) {
			j = 0
		}
	}
}
