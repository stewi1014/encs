package encodable_test

import (
	"bytes"
	"io"
	"testing"

	"github.com/stewi1014/encs/encodable"
)

// helper functions for testing

func checkSize(buff *bytes.Buffer, e encodable.Encodable, t *testing.T) {
	s := e.Size()
	if s < 0 {
		return
	}

	if buff.Len() > s {
		t.Fatalf("reported size smaller than written bytes; reported %v but wrote %v bytes", s, buff.Len())
	}
}

type buffer struct {
	buff []byte
	off  int
}

// reset resets reading
func (b *buffer) Reset() {
	b.off = 0
}

func (b *buffer) Read(buff []byte) (int, error) {
	n := copy(buff, b.buff[b.off:])
	b.off += n
	if n < len(buff) {
		return n, io.EOF
	}
	return n, nil
}

func (b *buffer) Write(buff []byte) (int, error) {
	copy(b.buff[b.grow(len(buff)):], buff)
	return len(buff), nil
}

func (b *buffer) grow(n int) int {
	l := len(b.buff)
	c := cap(b.buff)
	if l+n <= c {
		b.buff = b.buff[:l+n]
		return l
	}

	nb := make([]byte, l+n, c*2+n)
	copy(nb, b.buff)
	b.buff = nb
	return l
}
