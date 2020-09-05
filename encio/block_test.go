package encio_test

import (
	"bytes"
	"errors"
	"io"
	"io/ioutil"
	"math/rand"
	"testing"

	"github.com/maxatome/go-testdeep/td"
	"github.com/stewi1014/encs/encio"
)

func TestBlock(t *testing.T) {
	buff := new(bytes.Buffer)

	reader := encio.NewBlockReader(buff)
	writer := encio.NewBlockWriter(buff)

	maxlen := 100000

	rng := rand.New(rand.NewSource(256))
	send := make([]byte, maxlen)
	receive := make([]byte, maxlen)

	for i := 0; i < 10; i++ {
		l := rng.Intn(maxlen-1) + 1
		send = send[:l]
		for j := 0; j < l; j++ {
			send[j] = byte(rng.Uint32())
		}

		if err := encio.Write(send, writer); err != nil {
			t.Error(err)
		}

		n, err := reader.Read(receive)
		if err != nil && !errors.Is(err, io.EOF) {
			t.Error(err)
		}

		if !td.Cmp(t, receive[:n], send) {
			t.Errorf("Got: %v\n Wanted: %v", receive[:n], send)
		}

		if buff.Len() != 0 {
			t.Errorf("data remining in buffer %v", buff.Bytes())
		}
	}
}

func BenchmarkBlockWrite(b *testing.B) {
	l := 32

	writer := encio.NewBlockWriter(ioutil.Discard)
	send := make([]byte, l)

	rng := rand.New(rand.NewSource(256))
	for j := 0; j < l; j++ {
		send[j] = byte(rng.Uint32())
	}

	for i := 0; i < b.N; i++ {
		if err := encio.Write(send, writer); err != nil {
			b.Error(err)
		}
	}
}

func BenchmarkBlockRead(b *testing.B) {
	l := 32

	buff := new(encio.Buffer)

	writer := encio.NewBlockWriter(buff)
	send := make([]byte, l)

	rng := rand.New(rand.NewSource(256))
	for j := 0; j < l; j++ {
		send[j] = byte(rng.Uint32())
	}

	if err := encio.Write(send, writer); err != nil {
		b.Error(err)
	}

	reader := encio.NewBlockReader(encio.NewRepeatReader(*buff))

	for i := 0; i < b.N; i++ {
		if err := encio.Read(send, reader); err != nil {
			b.Error(err)
		}
	}
}
