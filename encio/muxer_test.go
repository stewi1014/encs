package encio_test

import (
	"hash/crc32"
	"io"
	"io/ioutil"
	"math/rand"
	"sync"
	"testing"

	"github.com/stewi1014/encs/encio"
)

func TestMuxer(t *testing.T) {
	t.Run("concurrent", testMuxConcurrent)
	t.Run("pipe", testMuxPipe)
}

func testMuxConcurrent(t *testing.T) {
	pr, pw := io.Pipe()

	streamCount := 60
	payloads := 100
	maxLen := 100

	t.Run("read", func(t *testing.T) {
		t.Parallel()
		r := encio.NewMuxReader(pr)
		var wg sync.WaitGroup
		wg.Add(streamCount)

		var id encio.UUID

		for i := 0; i < streamCount; i++ {
			err := encio.Read(id[:], r)
			if err != nil {
				t.Error(err)
			}

			s, err := r.Open(id)
			if err != nil {
				t.Fatal(err)
			}

			go func(s io.ReadCloser, seed int) {
				if err := testRead(s, seed, payloads, maxLen); err != nil {
					t.Error(err)
				}
				s.Close()
				wg.Done()
			}(s, i)
		}

		wg.Wait()
		r.Close()
	})

	t.Run("write", func(t *testing.T) {
		t.Parallel()
		w := encio.NewMuxWriter(pw)
		var wg sync.WaitGroup
		wg.Add(streamCount)

		for i := 0; i < streamCount; i++ {
			id := encio.NewUUID()
			n, err := w.Write(id[:])
			if n != len(encio.UUID{}) && err != nil {
				t.Error("writing stream id", n, err)
			}

			s, err := w.OpenStream(id)
			if err != nil {
				t.Fatal(err)
			}

			go func(s io.WriteCloser, seed int) {
				if err := testWrite(s, seed, payloads, maxLen); err != nil {
					t.Error(err)
				}
				s.Close()
				wg.Done()
			}(s, i)
		}

		wg.Wait()
		w.Close()
		pw.Close()
	})
}

func testMuxPipe(t *testing.T) {
	pr, pw := io.Pipe()

	count := 20
	maxLen := 20

	done := make(chan uint32)

	t.Run("writes", func(t *testing.T) {
		t.Parallel()

		rng := rand.New(rand.NewSource(0))
		hasher := crc32.NewIEEE()
		w := encio.NewMuxWriter(pw)

		for i := 0; i < count; i++ {
			buff := randomBytes(rng, maxLen)

			if err := encio.Write(buff, w); err != nil {
				t.Fatal(err)
			}

			_, err := hasher.Write(buff)
			if err != nil {
				panic(err)
			}
		}

		pw.Close()
		done <- hasher.Sum32()
	})

	t.Run("read", func(t *testing.T) {
		t.Parallel()

		r := encio.NewMuxReader(pr)

		buff := make([]byte, maxLen)
		hasher := crc32.NewIEEE()

		for {
			n, err := r.Read(buff[:maxLen])
			if err != nil {
				break
			}

			_, err = hasher.Write(buff[:n])
			if err != nil {
				panic(err)
			}
		}

		wanted := <-done
		if wanted != hasher.Sum32() {
			t.Fatal("wrong hash")
		}
	})
}

func BenchmarkMuxWrite(b *testing.B) {
	l := 256

	writer := encio.NewMuxWriter(ioutil.Discard)
	send := make([]byte, l)

	rng := rand.New(rand.NewSource(256))
	rng.Read(send)

	for i := 0; i < (b.N / l); i++ {
		if err := encio.Write(send, writer); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkMuxRead(b *testing.B) {
	l := 256

	buff := new(encio.Buffer)

	writer := encio.NewMuxWriter(buff)
	send := make([]byte, l)

	rng := rand.New(rand.NewSource(256))
	rng.Read(send)

	if err := encio.Write(send, writer); err != nil {
		b.Error(err)
	}

	reader := encio.NewMuxReader(encio.NewRepeatReader(*buff))

	for i := 0; i < (b.N / l); i++ {
		if err := encio.Read(send, reader); err != nil {
			b.Fatal(err)
		}
	}
}
