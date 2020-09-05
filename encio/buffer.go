package encio

import (
	"io"
)

// Buffer is a byte slice with helpful methods.
type Buffer []byte

// Write implements io.Writer.
// It appends buff to the buffer.
func (b *Buffer) Write(buff []byte) (int, error) {
	copy((*b)[b.Grow(len(buff)):], buff)
	return len(buff), nil
}

// Grow grows the buffer by n bytes, returning the previous length of the buffer.
func (b *Buffer) Grow(n int) int {
	buff := *b
	l := len(buff)
	c := cap(buff)
	if l+n <= c {
		*b = buff[:l+n]
		return l
	}

	*b = make([]byte, l+n, c*2+n)
	copy(*b, buff)
	return l
}

// NewRepeatReader return a RepeatReader that repeatedly reads from buff.
func NewRepeatReader(buff []byte) *RepeatReader {
	return &RepeatReader{
		buff: buff,
	}
}

// RepeatReader provides a method for repeatedly reading the same piece of data.
type RepeatReader struct {
	buff []byte
	off  int
}

// Read implements io.Reader.
// It reads from the internal buffer. When the end of the buffer is reached, it return io.EOF.
// Subsequent calls will begin reading from the beginning again.
func (r *RepeatReader) Read(buff []byte) (int, error) {
	n := copy(buff, r.buff)
	r.off += n
	if r.off == len(r.buff) {
		r.off = 0
		return n, io.EOF
	}
	return n, nil
}
