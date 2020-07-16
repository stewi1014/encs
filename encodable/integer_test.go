package encodable_test

import (
	"bytes"
	"fmt"
	"math/bits"
	"testing"
	"unsafe"

	"github.com/stewi1014/encs/encodable"
)

func TestUint8(t *testing.T) {
	enc := encodable.NewUint8()
	for tC := uint8(0); ; tC++ {
		testEqual(&tC, &tC, enc, t)

		if tC == 1<<8-1 {
			break
		}
	}
}

func TestUint16(t *testing.T) {
	enc := encodable.NewUint16()
	for tC := uint16(0); ; tC++ {
		testEqual(&tC, &tC, enc, t)

		if tC == 1<<16-1 {
			break
		}
	}
}

func TestUint32(t *testing.T) {
	testCases := []uint32{
		0, 1, 2, 3, 4, 5, 6, 254, 255, 256, 1<<32 - 1,
	}

	enc := encodable.NewUint32()

	for _, tC := range testCases {
		t.Run(fmt.Sprint(tC), func(t *testing.T) {
			testEqual(&tC, &tC, enc, t)
		})
	}
}

func TestUint64(t *testing.T) {
	testCases := []uint64{
		0, 1, 2, 3, 4, 5, 6, 254, 255, 256, 1<<32 - 1, 1<<64 - 1,
	}

	enc := encodable.NewUint64()

	for _, tC := range testCases {
		t.Run(fmt.Sprint(tC), func(t *testing.T) {
			testEqual(&tC, &tC, enc, t)
		})
	}
}

func TestUint(t *testing.T) {
	testCases := []uint{
		0, 1, 2, 3, 4, 5, 6, 254, 255, 256, 1<<bits.UintSize - 1,
	}

	enc := encodable.NewUint()

	for _, tC := range testCases {
		t.Run(fmt.Sprint(tC), func(t *testing.T) {
			testEqual(&tC, &tC, enc, t)
		})
	}
}

func TestInt8(t *testing.T) {
	enc := encodable.NewInt8()
	for tC := int8(-1 << 7); ; tC++ {
		testEqual(&tC, &tC, enc, t)

		if tC == 1<<7-1 {
			break
		}
	}
}

func TestInt16(t *testing.T) {
	enc := encodable.NewInt16()
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

	enc := encodable.NewInt32()

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

	enc := encodable.NewInt64()

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

	enc := encodable.NewInt()

	for _, tC := range testCases {
		t.Run(fmt.Sprint(tC), func(t *testing.T) {
			testEqual(&tC, &tC, enc, t)
		})
	}
}

func BenchmarkUint64(b *testing.B) {
	uints := []uint64{
		0, 1, 2, 3, 4, 5, 6, 254, 255, 256, 1<<32 - 1, 1<<64 - 1,
	}

	enc := encodable.NewUint64()
	buff := new(bytes.Buffer)
	j := 0
	var u uint64

	for i := 0; i < b.N; i++ {
		err := enc.Encode(unsafe.Pointer(&uints[j]), buff)
		if err != nil {
			b.Fatal(err)
		}
		err = enc.Decode(unsafe.Pointer(unsafe.Pointer(&u)), buff)
		if err != nil {
			b.Fatal(err)
		}
		if buff.Len() != 0 {
			b.Fatalf("data remaining in buffer %v", buff.Bytes())
		}
		j++
		if j >= len(uints) {
			j = 0
		}
	}
}

func BenchmarkUint(b *testing.B) {
	uints := []uint{
		0, 1, 2, 3, 4, 5, 6, 254, 255, 256, 1<<32 - 1,
	}

	enc := encodable.NewUint()
	buff := new(bytes.Buffer)
	j := 0
	var u uint

	for i := 0; i < b.N; i++ {
		err := enc.Encode(unsafe.Pointer(&uints[j]), buff)
		if err != nil {
			b.Fatal(err)
		}
		err = enc.Decode(unsafe.Pointer(unsafe.Pointer(&u)), buff)
		if err != nil {
			b.Fatal(err)
		}
		j++
		if j >= len(uints) {
			j = 0
		}
	}
}

func BenchmarkInt(b *testing.B) {
	ints := []int{
		0, 1, 2, 3, 4, 5, 6, 254, 255, 256, 1<<32 - 1, 1<<(bits.UintSize-1) - 1, -1, -1 << 63,
	}

	enc := encodable.NewInt()
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

func TestUintptr(t *testing.T) {
	testCases := []uintptr{
		0, 1, 2, 3, 4, 5, 6, 254, 255, 256, 1<<32 - 1, 1<<64 - 1,
	}

	enc := encodable.NewUintptr()

	for _, tC := range testCases {
		t.Run(fmt.Sprint(tC), func(t *testing.T) {
			testEqual(&tC, &tC, enc, t)
		})
	}
}
