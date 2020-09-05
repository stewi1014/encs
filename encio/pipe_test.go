package encio_test

import (
	"errors"
	"io"
	"io/ioutil"
	"math/rand"
	"testing"

	"github.com/stewi1014/encs/encio"
)

func TestPipe(t *testing.T) {
	rng := rand.New(rand.NewSource(0))
	count := 20
	maxBuffer := 500

	p := encio.Pipe()

	read := make(chan int64)
	var written int64

	r := encio.NewChecksumReader(p, nil)
	w := encio.NewChecksumWriter(p, nil)

	go t.Run("Pipe Read", func(t *testing.T) {
		n, err := io.Copy(ioutil.Discard, r)
		read <- n
		if !errors.Is(err, io.ErrClosedPipe) {
			t.Fatal(err)
		}
	})

	for i := 0; i < count; i++ {
		buff := make([]byte, rng.Intn(maxBuffer-1)+1)
		_, err := rng.Read(buff)
		if err != nil {
			panic(err)
		}

		n, err := w.Write(buff)
		if err != nil {
			t.Fatal(err)
		}

		written += int64(n)
	}

	p.Close()

	totalRead := <-read

	if written != totalRead {
		t.Fatalf("wrote %v bytes, but read %v", written, totalRead)
	}
}
