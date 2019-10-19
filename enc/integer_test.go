package enc_test

import (
	"bytes"
	"fmt"
	"math/bits"
	"testing"
	"unsafe"

	"github.com/stewi1014/encs/enc"
)

var (
	intSink     int
	int8Sink    int8
	int16Sink   int16
	int32Sink   int32
	int64Sink   int64
	uintSink    uint
	uint8Sink   uint8
	uint16Sink  uint16
	uint32Sink  uint32
	uint64Sink  uint64
	uintptrSink uintptr
)

func TestUint8(t *testing.T) {
	e := enc.NewUint8()
	buff := new(bytes.Buffer)
	for i := uint8(0); ; i++ {
		err := e.Encode(unsafe.Pointer(&i), buff)
		if err != nil {
			t.Fatalf("Error encoding: %v", err)
		}
		var d uint8
		err = e.Decode(unsafe.Pointer(&d), buff)
		if err != nil {
			t.Fatalf("Error decoding: %v", err)
		}

		if d != i {
			t.Fatalf("Encoded %v, but got %v", i, d)
		}

		if buff.Len() != 0 {
			t.Fatalf("Data remaining in buffer: %v", buff.Bytes())
		}

		if i == 1<<8-1 {
			break
		}
	}
}

func TestUint16(t *testing.T) {
	e := enc.NewUint16()
	buff := new(bytes.Buffer)
	for i := uint16(0); ; i++ {
		err := e.Encode(unsafe.Pointer(&i), buff)
		if err != nil {
			t.Fatalf("Error encoding: %v", err)
		}
		var d uint16
		err = e.Decode(unsafe.Pointer(&d), buff)
		if err != nil {
			t.Fatalf("Error decoding: %v", err)
		}

		if d != i {
			t.Fatalf("Encoded %v, but got %v", i, d)
		}

		if buff.Len() != 0 {
			t.Fatalf("Data remaining in buffer: %v", buff.Bytes())
		}

		if i == 1<<16-1 {
			break
		}
	}
}

func TestUint32(t *testing.T) {
	testCases := []uint32{
		0, 1, 2, 3, 4, 5, 6, 254, 255, 256, 1<<32 - 1,
	}

	e := enc.NewUint32()
	buff := new(bytes.Buffer)

	for _, tC := range testCases {
		t.Run(fmt.Sprint(tC), func(t *testing.T) {
			err := e.Encode(unsafe.Pointer(&tC), buff)
			if err != nil {
				t.Fatalf("Encode error: %v", err)
			}

			var d uint32
			err = e.Decode(unsafe.Pointer(&d), buff)
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

func TestUint64(t *testing.T) {
	testCases := []uint64{
		0, 1, 2, 3, 4, 5, 6, 254, 255, 256, 1<<32 - 1, 1<<64 - 1,
	}

	e := enc.NewUint64()
	buff := new(bytes.Buffer)

	for _, tC := range testCases {
		t.Run(fmt.Sprint(tC), func(t *testing.T) {
			err := e.Encode(unsafe.Pointer(&tC), buff)
			if err != nil {
				t.Fatalf("Encode error: %v", err)
			}

			var d uint64
			err = e.Decode(unsafe.Pointer(&d), buff)
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

func TestUint(t *testing.T) {
	testCases := []uint{
		0, 1, 2, 3, 4, 5, 6, 254, 255, 256, 1<<bits.UintSize - 1,
	}

	e := enc.NewUint()
	buff := new(bytes.Buffer)

	for _, tC := range testCases {
		t.Run(fmt.Sprint(tC), func(t *testing.T) {
			err := e.Encode(unsafe.Pointer(&tC), buff)
			if err != nil {
				t.Fatalf("Encode error: %v", err)
			}

			bytes := buff.Bytes()

			var d uint
			err = e.Decode(unsafe.Pointer(&d), buff)
			if err != nil {
				t.Fatalf("Decode error: %v", err)
			}

			if d != tC {
				t.Fatalf("Encoded %v, but got %v, buffer %v", tC, d, bytes)
			}

			if buff.Len() != 0 {
				t.Fatalf("Data remaining in buffer: %v", buff.Bytes())
			}
		})
	}
}

func TestInt8(t *testing.T) {
	e := enc.NewInt8()
	buff := new(bytes.Buffer)
	for i := int8(-1 << 7); ; i++ {
		err := e.Encode(unsafe.Pointer(&i), buff)
		if err != nil {
			t.Fatalf("Error encoding: %v", err)
		}
		var d int8
		err = e.Decode(unsafe.Pointer(&d), buff)
		if err != nil {
			t.Fatalf("Error decoding: %v", err)
		}

		if d != i {
			t.Fatalf("Encoded %v, but got %v", i, d)
		}
		if buff.Len() != 0 {
			t.Fatalf("Data remaining in buffer: %v", buff.Bytes())
		}
		if i == 1<<7-1 {
			break
		}
	}
}

func TestInt16(t *testing.T) {
	e := enc.NewInt16()
	buff := new(bytes.Buffer)
	for i := int16(-1 << 15); ; i++ {
		err := e.Encode(unsafe.Pointer(&i), buff)
		if err != nil {
			t.Fatalf("Error encoding: %v", err)
		}
		var d int16
		err = e.Decode(unsafe.Pointer(&d), buff)
		if err != nil {
			t.Fatalf("Error decoding: %v", err)
		}

		if d != i {
			t.Fatalf("Encoded %v, but got %v", i, d)
		}

		if buff.Len() != 0 {
			t.Fatalf("Data remaining in buffer: %v", buff.Bytes())
		}
		if i == 1<<15-1 {
			break
		}
	}
}

func TestInt32(t *testing.T) {
	testCases := []int32{
		0, 1, 2, 3, 4, 5, 6, 254, 255, 256, -1 << 31, 1<<31 - 1, -1,
	}

	e := enc.NewInt32()
	buff := new(bytes.Buffer)

	for _, tC := range testCases {
		t.Run(fmt.Sprint(tC), func(t *testing.T) {
			err := e.Encode(unsafe.Pointer(&tC), buff)
			if err != nil {
				t.Fatalf("Encode error: %v", err)
			}

			var d int32
			err = e.Decode(unsafe.Pointer(&d), buff)
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

func TestInt64(t *testing.T) {
	testCases := []int64{
		0, 1, 2, 3, 4, 5, 6, 254, 255, 256, 1<<32 - 1, 1<<63 - 1, -1, -1 << 63,
	}

	e := enc.NewInt64()
	buff := new(bytes.Buffer)

	for _, tC := range testCases {
		t.Run(fmt.Sprint(tC), func(t *testing.T) {
			err := e.Encode(unsafe.Pointer(&tC), buff)
			if err != nil {
				t.Fatalf("Encode error: %v", err)
			}

			var d int64
			err = e.Decode(unsafe.Pointer(&d), buff)
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

func TestInt(t *testing.T) {
	testCases := []int{
		0, 1, 2, 3, 4, 5, 6, 254, 255, 256, 1<<32 - 1, 1<<(bits.UintSize-1) - 1, -1, -1 << 63,
	}

	e := enc.NewInt()
	buff := new(bytes.Buffer)

	for _, tC := range testCases {
		t.Run(fmt.Sprint(tC), func(t *testing.T) {
			err := e.Encode(unsafe.Pointer(&tC), buff)
			if err != nil {
				t.Fatalf("Encode error: %v", err)
			}

			bytes := buff.Bytes()

			var d int
			err = e.Decode(unsafe.Pointer(&d), buff)
			if err != nil {
				t.Fatalf("Decode error: %v", err)
			}

			if d != tC {
				t.Fatalf("Encoded %v, but got %v, buffer: %v", tC, d, bytes)
			}

			if buff.Len() != 0 {
				t.Fatalf("Data remaining in buffer: %v", buff.Bytes())
			}
		})
	}
}

func BenchmarkUint64(b *testing.B) {
	uints := []uint64{
		0, 1, 2, 3, 4, 5, 6, 254, 255, 256, 1<<32 - 1, 1<<64 - 1,
	}

	enc := enc.NewUint64()
	buff := new(bytes.Buffer)
	j := 0
	var u uint64

	for i := 0; i < b.N; i++ {
		enc.Encode(unsafe.Pointer(&uints[j]), buff)
		enc.Decode(unsafe.Pointer(unsafe.Pointer(&u)), buff)
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

	enc := enc.NewUint()
	buff := new(bytes.Buffer)
	j := 0
	var u uint

	for i := 0; i < b.N; i++ {
		enc.Encode(unsafe.Pointer(&uints[j]), buff)
		enc.Decode(unsafe.Pointer(unsafe.Pointer(&u)), buff)
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

	enc := enc.NewInt()
	buff := new(bytes.Buffer)
	j := 0
	var u int

	for i := 0; i < b.N; i++ {
		enc.Encode(unsafe.Pointer(&ints[j]), buff)
		enc.Decode(unsafe.Pointer(unsafe.Pointer(&u)), buff)

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

	enc := enc.NewUintptr()
	buff := new(bytes.Buffer)

	for _, tC := range testCases {
		t.Run(fmt.Sprint(tC), func(t *testing.T) {
			err := enc.Encode(unsafe.Pointer(&tC), buff)
			if err != nil {
				t.Fatalf("Encode error: %v", err)
			}

			var d uintptr
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
