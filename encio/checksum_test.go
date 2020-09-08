package encio_test

import (
	"bytes"
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"errors"
	"hash"
	"hash/adler32"
	"hash/crc32"
	"hash/crc64"
	"hash/fnv"
	"math/rand"
	"reflect"
	"testing"

	"github.com/maxatome/go-testdeep/td"

	"github.com/stewi1014/encs/encio"
)

func TestChecksum(t *testing.T) {
	count := 20
	seed := int64(0)
	maxBuffer := 500

	hashers := []func() hash.Hash{
		// hash.Hash32 or hash.Hash64 implementers.
		func() hash.Hash { return crc32.NewIEEE() },
		func() hash.Hash { return crc64.New(crc64.MakeTable(crc64.ISO)) },
		func() hash.Hash { return adler32.New() },
		func() hash.Hash { return fnv.New64() },
		func() hash.Hash { return fnv.New32() },

		fnv.New128,
		md5.New,
		sha256.New,
		sha512.New,
		sha1.New,
	}

	for _, hasher := range hashers {
		name := reflect.TypeOf(hasher()).String()
		t.Run(name, func(t *testing.T) {
			testChecksum(count, seed, maxBuffer, hasher, t)
		})
	}
}

func testChecksum(count int, seed int64, maxBuffer int, hasher func() hash.Hash, t *testing.T) {
	t.Run("Returns error on flipped bit", func(t *testing.T) {
		rng := rand.New(rand.NewSource(seed))

		var send encio.Buffer
		var receive encio.Buffer

		for i := 0; i < count; i++ {
			buff := new(bytes.Buffer)

			cr := encio.NewChecksumReader(buff, hasher())
			cw := encio.NewChecksumWriter(buff, hasher())

			size := rng.Intn(maxBuffer-1) + 1

			send = send[:0]
			send.Grow(size)
			receive = receive[:0]
			receive.Grow(size)

			if err := encio.Read(send, rng); err != nil {
				t.Fatal(err)
			}

			if err := encio.Write(send, cw); err != nil {
				t.Fatal(err)
			}

			var flippedIndex int
			if size <= 1 {
				flippedIndex = 0
			} else {
				flippedIndex = rand.Intn(size - 1)
			}

			buff.Bytes()[flippedIndex] = buff.Bytes()[flippedIndex] + 1

			err := encio.Read(receive, cr)
			if !errors.Is(err, encio.ErrMalformed) {
				t.Fatalf("Wanted %v, but got %v. Bit flipped at %v", encio.ErrMalformed, err, flippedIndex)
			}
		}
	})

	t.Run("Returns error on out-of-order data", func(t *testing.T) {
		rng := rand.New(rand.NewSource(seed))

		var b encio.Buffer
		var receive encio.Buffer

		for i := 0; i < count; i++ {
			buff := new(bytes.Buffer)

			cr := encio.NewChecksumReader(buff, hasher())
			cw := encio.NewChecksumWriter(buff, hasher())

			size := rng.Intn(maxBuffer-1) + 1

			b = b[:0]
			b.Grow(size)
			receive = receive[:0]
			receive.Grow(size)

			if err := encio.Read(b, rng); err != nil {
				t.Fatal(err)
			}

			if err := encio.Write(b, cw); err != nil {
				t.Fatal(err)
			}

			// Steal the first bit of data.
			var first encio.Buffer
			_, err := first.ReadFrom(buff)
			if err != nil {
				t.Fatalf("copying buffer for the test gave %v", err)
			}

			if err := encio.Write(b, cw); err != nil {
				t.Fatal(err)
			}

			// Put the first packet back
			buff.Write(first)

			err = encio.Read(receive, cr)
			if !errors.Is(err, encio.ErrMalformed) {
				t.Fatalf("Wanted %v, but got %v", encio.ErrMalformed, err)
			}
		}
	})

	rng := rand.New(rand.NewSource(seed))

	var b encio.Buffer
	var receive encio.Buffer

	buff := new(bytes.Buffer)

	cr := encio.NewChecksumReader(buff, hasher())
	cw := encio.NewChecksumWriter(buff, hasher())

	for i := 0; i < count; i++ {
		size := rng.Intn(maxBuffer)

		b = b[:0]
		b.Grow(size)
		receive = receive[:0]
		receive.Grow(size)

		if err := encio.Read(b, rng); err != nil {
			t.Fatal(err)
		}

		if err := encio.Write(b, cw); err != nil {
			t.Fatal(err)
		}

		if err := encio.Read(receive, cr); err != nil {
			t.Fatal(err)
		}

		td.Cmp(t, receive, b)
	}
}