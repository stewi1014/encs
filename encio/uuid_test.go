package encio_test

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/stewi1014/encs/encio"
)

func TestUUID(t *testing.T) {
	testCases := [][2]uint64{
		{0, 0},
		{1, 1},
		{2, 2},
		{3, 3},
		{4, 4},
		{246, 246},
		{247, 247},
		{248, 248},
		{249, 249},
		{250, 250},
		{251, 251},
		{252, 252},
		{253, 253},
		{254, 254},
		{255, 255},
		{256, 256},
		{257, 257},
		{1 << 8, 1 << 8},
		{1 << 16, 1 << 16},
		{1 << 24, 1 << 24},
		{1 << 32, 1 << 32},
		{1 << 40, 1 << 40},
		{1 << 48, 1 << 48},
		{1 << 56, 1 << 56},
		{1<<64 - 1, 1<<64 - 1},
	}

	enc := encio.UUID{}

	for _, tC := range testCases {
		t.Run(fmt.Sprint(tC), func(t *testing.T) {
			buff := new(bytes.Buffer)

			if err := enc.EncodeUUID(buff, tC); err != nil {
				t.Fatal(err)
			}

			id, err := enc.DecodeUUID(buff)
			if err != nil {
				t.Fatal(err)
			}

			if id != tC {
				t.Fatalf("Wrong number, wanted: %v, got %v", tC, id)
			}

			if buff.Len() != 0 {
				t.Fatalf("data remaining in buffer %v", buff.Bytes())
			}
		})
	}
}
