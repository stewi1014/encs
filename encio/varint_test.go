package encio_test

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/stewi1014/encs/encio"
)

func TestUvarint(t *testing.T) {
	testCases := []uint32{
		0, 1, 2, 3, 4,
		246, 247, 248, 249, 250, 251, 252, 253, 254, 255, 256, 257,
		1 << 8, 1 << 16, 1 << 24, 1<<32 - 1,
	}

	enc := encio.Uvarint{}

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
