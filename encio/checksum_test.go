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
	"io"
	"math/rand"
	"reflect"
	"testing"

	"github.com/maxatome/go-testdeep/td"

	"github.com/stewi1014/encs/encio"
)

func TestChecksum(t *testing.T) {
	count := 20
	seed := 0
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

func testChecksum(count, seed, maxBuffer int, hasher func() hash.Hash, t *testing.T) {
	rng := rand.New(rand.NewSource(int64(seed)))

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

	t.Run("Returns error on flipped bit", func(t *testing.T) {
		testChecksumBadData(t, count, seed, maxBuffer, hasher)
	})

	t.Run("Returns error on out-of-order data", func(t *testing.T) {
		testChecksumOutOfOrder(t, count, seed, maxBuffer, hasher)
	})

	t.Run("can continue after error", func(t *testing.T) {
		testChecksumContinue(t, seed, hasher)
	})
}

func testChecksumBadData(t *testing.T, count, seed, maxBuffer int, hasher func() hash.Hash) {
	rng := rand.New(rand.NewSource(int64(seed)))

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
}

func testChecksumOutOfOrder(t *testing.T, count, seed, maxBuffer int, hasher func() hash.Hash) {
	rng := rand.New(rand.NewSource(int64(seed)))

	var wbuff encio.Buffer
	cw := encio.NewChecksumWriter(&wbuff, hasher())

	send := make([][]byte, count)
	buffs := make([]encio.Buffer, count)

	for i := range send {
		send[i] = make([]byte, maxBuffer)
		rng.Read(send[i])

		n, err := cw.Write(send[i])
		if err != nil {
			t.Error(err)
		}
		if n != len(send[i]) {
			t.Error(io.ErrShortWrite)
		}

		buffs[i].MWrite(wbuff)
		wbuff.Reset()
	}

	b := new(encio.ReadBuffer)
	b.MWrite(buffs[0])
	b.MWrite(buffs[2])
	b.MWrite(buffs[1])

	rbuff := make(encio.Buffer, maxBuffer)
	cr := encio.NewChecksumReader(b, hasher())

	n, err := cr.Read(rbuff)
	if err != nil {
		t.Error(err)
	}
	if n != len(rbuff) {
		t.Error(io.ErrUnexpectedEOF)
	}

	td.Cmp(t, []byte(rbuff), send[0])

	n, err = cr.Read(rbuff)
	if err == nil {
		t.Error("no error on out of order data", n)
	}

	n, err = cr.Read(rbuff)
	if err != nil {
		t.Error(err)
	}
	if n != len(rbuff) {
		t.Error(io.ErrUnexpectedEOF)
	}

	td.Cmp(t, []byte(rbuff), send[1])
}

func testChecksumContinue(t *testing.T, seed int, hasher func() hash.Hash) {
	rng := rand.New(rand.NewSource(int64(seed)))

	size := 50

	send := make([][]byte, 3)
	for i := range send {
		send[i] = make([]byte, size)
		rng.Read(send[i])
	}

	b := new(bytes.Buffer)

	var wbuff encio.Buffer
	cw := encio.NewChecksumWriter(&wbuff, hasher())

	n, err := cw.Write(send[0])
	if err != nil {
		t.Error(err)
	}
	if n != len(send[0]) {
		t.Error(io.ErrShortWrite)
	}
	b.Write(wbuff)
	wbuff.Reset()

	n, err = cw.Write(send[1])
	if err != nil {
		t.Error(err)
	}
	if n != len(send[1]) {
		t.Error(io.ErrShortWrite)
	}
	wbuff[0]++
	b.Write(wbuff)
	wbuff.Reset()

	n, err = cw.Write(send[2])
	if err != nil {
		t.Error(err)
	}
	if n != len(send[2]) {
		t.Error(io.ErrShortWrite)
	}
	b.Write(wbuff)
	wbuff.Reset()

	rbuff := make(encio.Buffer, size)
	cr := encio.NewChecksumReader(b, hasher())

	n, err = cr.Read(rbuff)
	if err != nil {
		t.Error(err)
	}
	if n != len(rbuff) {
		t.Error(io.ErrUnexpectedEOF)
	}

	td.Cmp(t, []byte(rbuff), send[0])

	n, err = cr.Read(rbuff)
	if err == nil {
		t.Error("no error on damaged data", n)
	}

	n, err = cr.Read(rbuff)
	if err != nil {
		t.Error(err)
	}
	if n != len(rbuff) {
		t.Error(io.ErrUnexpectedEOF)
	}

	td.Cmp(t, []byte(rbuff), send[2])
}
