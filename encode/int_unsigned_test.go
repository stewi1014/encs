package encode_test

import (
	"bytes"
	"fmt"
	"math/bits"
	"reflect"
	"testing"
	"unsafe"

	"github.com/stewi1014/encs/encode"
)

func TestUint8(t *testing.T) {
	enc := encode.NewUint8(reflect.TypeOf(uint8(0)))
	for tC := uint8(0); ; tC++ {
		testEqual(&tC, &tC, enc, t)

		if tC == 1<<8-1 {
			break
		}
	}
}

func TestUint16(t *testing.T) {
	enc := encode.NewUint16(reflect.TypeOf(uint16(0)))
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

	enc := encode.NewUint32(reflect.TypeOf(uint32(0)))

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

	enc := encode.NewUint64(reflect.TypeOf(uint64(0)))

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

	enc := encode.NewUint(reflect.TypeOf(uint(0)))

	for _, tC := range testCases {
		t.Run(fmt.Sprint(tC), func(t *testing.T) {
			testEqual(&tC, &tC, enc, t)
		})
	}
}

func TestUintptr(t *testing.T) {
	testCases := []uintptr{
		0, 1, 2, 3, 4, 5, 6, 254, 255, 256, 1<<32 - 1, 1<<64 - 1,
	}

	enc := encode.NewUintptr(reflect.TypeOf(uintptr(0)))

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

	enc := encode.NewUint64(reflect.TypeOf(uint64(0)))
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

	enc := encode.NewUint(reflect.TypeOf(uint(0)))
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
