package encio

import (
	"errors"
	"fmt"
	"io"
)

// Buffer is a byte slice with helpful methods.
type Buffer []byte

// Write implements io.Writer.
// It appends buff to the buffer.
func (b *Buffer) Write(buff []byte) (int, error) {
	n := copy((*b)[b.Grow(len(buff)):], buff)
	return n, nil
}

// ReadFrom implements io.ReaderFrom.
// It reads from r, filling the buffer until either an error is returned by r, or no progress is made.
func (b *Buffer) ReadFrom(r io.Reader) (read int64, err error) {
	take := 256
	var n int

	for {
		l := b.Grow(take)
		n, err = r.Read((*b)[l:])

		*b = (*b)[:l+n]
		read += int64(n)

		if err != nil {
			break
		}

		if n == 0 {
			return read, NewIOError(
				io.ErrNoProgress,
				r,
				fmt.Sprintf(
					"trying to read %v bytes",
					take,
				),
				0,
			)
		}

		take *= 2
		if uintptr(take) > TooBig {
			return read, NewIOError(
				errors.New("too big"),
				r,
				fmt.Sprintf(
					"refusing to buffer %v bytes as it's too big",
					take,
				),
				0,
			)
		}
	}

	if err == io.EOF {
		err = nil
	}
	return
}

// Len returns the length of the buffer.
func (b *Buffer) Len() int {
	if b == nil {
		return 0
	}
	return len(*b)
}

// Cap return shte capacity of the buffer.
func (b *Buffer) Cap() int {
	if b == nil {
		return 0
	}
	return cap(*b)
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
	n := copy(buff, r.buff[r.off:])
	r.off += n
	if r.off == len(r.buff) {
		r.off = 0
		return n, io.EOF
	}
	return n, nil
}
