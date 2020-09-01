package encio_test

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"testing"

	"github.com/stewi1014/encs/encio"
)

func TestUint32(t *testing.T) {
	testCases := []uint32{
		0, 1, 2, 3, 4,
		246, 247, 248, 249, 250, 251, 252, 253, 254, 255, 256, 257,
		1 << 8, 1 << 16, 1 << 24, 1<<32 - 1,
	}

	enc := encio.NewUint32()

	for _, tC := range testCases {
		t.Run(fmt.Sprint(tC), func(t *testing.T) {
			buff := new(bytes.Buffer)

			if err := enc.Encode(buff, tC); err != nil {
				t.Fatal(err)
			}

			n, err := enc.Decode(buff)
			if err != nil {
				t.Fatal(err)
			}

			if n != tC {
				t.Fatalf("Wrong number, wanted: %v, got %v", tC, n)
			}

			if buff.Len() != 0 {
				t.Fatalf("data remaining in buffer %v", buff.Bytes())
			}
		})
	}
}

func TestInt32(t *testing.T) {
	testCases := []int32{
		0, 1, 2, 3, 4,
		246, 247, 248, 249, 250, 251, 252, 253, 254, 255, 256, 257,
		1 << 8, 1 << 16, 1 << 24, -1 << 31, -1,
	}

	enc := encio.NewInt32()

	for _, tC := range testCases {
		t.Run(fmt.Sprint(tC), func(t *testing.T) {
			buff := new(bytes.Buffer)

			if err := enc.Encode(buff, tC); err != nil {
				t.Fatal(err)
			}

			n, err := enc.Decode(buff)
			if err != nil {
				t.Fatal(err)
			}

			if n != tC {
				t.Fatalf("Wrong number, wanted: %v, got %v", tC, n)
			}

			if buff.Len() != 0 {
				t.Fatalf("data remaining in buffer %v", buff.Bytes())
			}
		})
	}
}

func TestVarint(t *testing.T) {
	testCases := []uint32{
		0, 1, 2, 3, 4,
		246, 247, 248, 249, 250, 251, 252, 253, 254, 255, 256, 257,
		1 << 8, 1 << 16, 1 << 24, 1<<30 - 1,
	}

	enc := encio.NewVaruint32()

	for _, tC := range testCases {
		t.Run(fmt.Sprint(tC), func(t *testing.T) {
			buff := new(bytes.Buffer)

			if _, err := enc.Encode(buff, tC); err != nil {
				t.Fatal(err)
			}

			n, err := enc.Decode(buff)
			if err != nil {
				t.Fatal(err)
			}

			if n != tC {
				t.Fatalf("Wrong number, wanted: %v, got %v", tC, n)
			}

			if buff.Len() != 0 {
				t.Fatalf("data remaining in buffer %v", buff.Bytes())
			}
		})
	}
}

func BenchmarkUintEncode(b *testing.B) {
	enc := encio.NewUint32()
	for i := 0; i < b.N; i++ {
		if err := enc.Encode(ioutil.Discard, uint32(i)); err != nil {
			panic(err)
		}
	}
}

func BenchmarkIntEncode(b *testing.B) {
	enc := encio.NewInt32()
	for i := 0; i < b.N; i++ {
		if err := enc.Encode(ioutil.Discard, int32(i)); err != nil {
			panic(err)
		}
	}
}
func BenchmarkVarintEncode(b *testing.B) {
	enc := encio.NewVaruint32()
	for i := 0; i < b.N; i++ {
		if _, err := enc.Encode(ioutil.Discard, uint32(i)); err != nil {
			panic(err)
		}
	}
}
