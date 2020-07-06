package encio_test

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/stewi1014/encs/encio"
)

func TestUint(t *testing.T) {
	testCases := []uint32{
		0, 1, 2, 3, 4,
		246, 247, 248, 249, 250, 251, 252, 253, 254, 255, 256, 257,
		1 << 8, 1 << 16, 1 << 24, 1<<32 - 1,
	}

	enc := encio.Uint{}

	for _, tC := range testCases {
		t.Run(fmt.Sprint(tC), func(t *testing.T) {
			buff := new(bytes.Buffer)

			if err := enc.EncodeUint32(buff, tC); err != nil {
				t.Fatal(err)
			}

			n, err := enc.DecodeUint32(buff)
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

func TestInt(t *testing.T) {
	testCases := []int32{
		0, 1, 2, 3, 4,
		246, 247, 248, 249, 250, 251, 252, 253, 254, 255, 256, 257,
		1 << 8, 1 << 16, 1 << 24, -1 << 31, -1,
	}

	enc := encio.Int{}

	for _, tC := range testCases {
		t.Run(fmt.Sprint(tC), func(t *testing.T) {
			buff := new(bytes.Buffer)

			if err := enc.EncodeInt32(buff, tC); err != nil {
				t.Fatal(err)
			}

			n, err := enc.DecodeInt32(buff)
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
