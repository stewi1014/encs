package encio

import (
	"errors"
	"fmt"
	"io"
)

// Buffer is a byte slice with helpful methods.
type Buffer []byte

// Write implements io.Writer, appending buff to the buffer.
// It cannot fail, and always returns len(buff) and nil.
func (b *Buffer) Write(buff []byte) (int, error) {
	b.MWrite(buff)
	return len(buff), nil
}

// MWrite is the same as Write, but has no return values.
func (b *Buffer) MWrite(buff []byte) {
	copy((*b)[b.Grow(len(buff)):], buff)
}

// WriteByte implements io.ByteWriter.
// It cannot fail, and always returns nil.
func (b *Buffer) WriteByte(by byte) error {
	b.MWriteByte(by)
	return nil
}

// MWriteByte is the same as WriteByte, but has no return values.
func (b *Buffer) MWriteByte(by byte) {
	(*b)[b.Grow(1)] = by
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
				ErrMalformed,
				r,
				fmt.Sprintf(
					"refusing to buffer %v bytes as it's too big",
					take,
				),
				0,
			)
		}
	}

	if errors.Is(err, io.EOF) {
		err = nil
	}
	return
}

// ReadNFrom reads n bytes from r, appending to to the end of the buffer.
func (b *Buffer) ReadNFrom(r io.Reader, n int) (read int, err error) {
	if n == 0 {
		return 0, nil
	}

	start := b.Grow(n)

	m := -1
	for read < n && err == nil && m != 0 {
		m, err = r.Read((*b)[start+read:])
		read += m
	}

	if err == nil && m == 0 {
		err = io.ErrNoProgress
	}

	(*b) = (*b)[:start+read]
	return read, err
}

// Len returns the length of the buffer.
func (b *Buffer) Len() int {
	if b == nil {
		return 0
	}
	return len(*b)
}

// Cap returns the capacity of the buffer.
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

// Reset sets the buffer's length to 0, keeping it for later use.
func (b *Buffer) Reset() { *b = (*b)[:0] }

// ReadBuffer provides similar functionality to Buffer, but has an offset integer
// which can be used when reading to read from the buffer without de-allocating memory at the start of the slice.
type ReadBuffer struct {
	buff Buffer
	off  int
}

// Write implements io.Writer, appending buff to the buffer.
// It cannot fail and always returns len(buff) and nil.
func (b *ReadBuffer) Write(buff []byte) (int, error) {
	b.MWrite(buff)
	return len(buff), nil
}

// MWrite is the same as Write, except it has no return values.
func (b *ReadBuffer) MWrite(buff []byte) {
	copy(b.buff[b.grow(len(buff)):], buff)
}

// WriteByte implements io.ByteWriter.
// It cannot fail and always returns nil.
func (b *ReadBuffer) WriteByte(by byte) error {
	b.MWriteByte(by)
	return nil
}

// MWriteByte is the same as WriteByte, except it has no return values.
func (b *ReadBuffer) MWriteByte(by byte) {
	b.buff[b.grow(1)] = by
}

// MRead is the same as Read, except it returns no error.
// It cannot fail. If the buffer is empty, it returns 0.
func (b *ReadBuffer) MRead(buff []byte) int {
	n := copy(buff, b.buff[b.off:])
	b.off += n
	return n
}

// Read implements io.Reader.
// It cannot fail and always returns the smaller of the internal or given buffer length.
// If the end of the buffer is reached it returns io.EOF.
func (b *ReadBuffer) Read(buff []byte) (int, error) {
	n := b.MRead(buff)
	if n < len(buff) {
		return n, io.EOF
	}
	return n, nil
}

// ReadByte implements io.ByteReader.
// It returns the next byte or io.EOF if at the end of the buffer.
func (b *ReadBuffer) ReadByte() (byte, error) {
	if b.Len() < 1 {
		return 0, io.EOF
	}

	b.off++
	return b.buff[b.off-1], nil
}

// ReadBytes returns the next n bytes in the buffer.
// The returned buffer may be modified on subsequent calls.
func (b *ReadBuffer) ReadBytes(n int) (buff Buffer, err error) {
	if b.Len() < n {
		n = b.Len()
		err = io.EOF
	}
	buff = b.buff[b.off : b.off+n]
	b.off += n
	return
}

// Len returns the length of the unread portion of the buffer.
func (b *ReadBuffer) Len() int { return b.buff.Len() - b.off }

// Reset resets the buffer replacing the buffer with buff.
// It returns the internal buffer as it was before the reset.
// The returned buffer does not take into account the current read offset.
func (b *ReadBuffer) Reset(buff []byte) (old []byte) {
	old = b.buff
	b.buff = buff
	b.off = 0
	return
}

// Consume removes n bytes from the buffer.
func (b *ReadBuffer) Consume(n int) (err error) {
	if b.Len() < n {
		n = b.Len()
		err = io.EOF
	}

	b.off += n
	return
}

func (b *ReadBuffer) grow(n int) (w int) {
	l := len(b.buff)
	if cap(b.buff) >= l+n {
		// Have enough space
		b.buff = b.buff[:l+n]
		return l
	}

	if cap(b.buff) >= ((l+n)-b.off)*8 {
		// Slide, but only if we can reclaim a reasonably large amount of data.
		w = copy(b.buff, b.buff[b.off:])
		b.buff = b.buff[:l+n-b.off]
		b.off = 0
		return
	}

	// Allocate
	nb := make([]byte, (l-b.off)+n, (cap(b.buff)*4)+n)
	w = copy(nb, b.buff[b.off:])
	b.buff = nb
	b.off = 0
	return w
}

// NewRepeatReader returns a RepeatReader using buff as its data.
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
// It reads from the internal buffer. When the end of the buffer is reached, it returns io.EOF.
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
