package encio_test

import (
	"errors"
	"io"
	"io/ioutil"
	"math/rand"
	"testing"
	"time"

	"github.com/stewi1014/encs/encio"
)

func TestPipe(t *testing.T) {
	rng := rand.New(rand.NewSource(0))
	count := 100
	maxBuffer := 500

	pr, pw := encio.Pipe()

	read := make(chan int64)
	var written int64

	r := encio.NewChecksumReader(pr, nil)
	w := encio.NewChecksumWriter(pw, nil)

	go t.Run("Pipe Read", func(t *testing.T) {
		n, err := io.Copy(ioutil.Discard, r)
		read <- n
		if !errors.Is(err, io.ErrClosedPipe) {
			t.Fatal(err)
		}
	})

	for i := 0; i < count; i++ {
		// Add a small random delay so the synchronisation logic is tested better.
		time.Sleep(time.Duration(rng.Intn(10000)) * time.Microsecond)

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

	pw.Close()
	pr.Close()

	totalRead := <-read

	if written != totalRead {
		t.Fatalf("wrote %v bytes, but read %v", written, totalRead)
	}
}
