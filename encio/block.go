package encio

import (
	"fmt"
	"io"
)

const (
	// EscapeByte is the byte used to define the start of a block.
	// It should be as uncommon as possible.
	EscapeByte byte = 253

	// this must be an invalid length
	escapeEscape = 0
)

// NewBlockWriter returns a new BlockWriter writing to w.
func NewBlockWriter(w io.Writer) *BlockWriter {
	return &BlockWriter{
		w: w,
	}
}

// BlockWriter provides a method for writing data in defined payloads.
type BlockWriter struct {
	buff []byte
	w    io.Writer
}

// Write implements io.Writer.
// The buffer passed to Write is immediately written along with a 5 byte header.
func (e *BlockWriter) Write(buff []byte) (int, error) {
	l := len(buff)
	if l == 0 {
		return 0, nil
	}

	e.buff = e.buff[:0]
	e.grow(len(buff) + 5)
	c := copy(e.buff[5:], buff)

	e.escape(5, len(e.buff))

	n := EncodeVarUint32(e.buff, uint32(len(e.buff)-5))
	copy(e.buff[5-n:5], e.buff[0:n])
	e.buff[4-n] = EscapeByte

	e.escape(5-n, 5)

	return c, Write(e.buff[4-n:], e.w)
}

func (e *BlockWriter) escape(x, y int) {
	for i := x; i < y; i++ {
		if e.buff[i] == EscapeByte {
			e.insert(1, i+1)
			e.buff[i+1] = escapeEscape
			y++
		}
	}
}

func (e *BlockWriter) insert(size, index int) {
	l := e.grow(size)
	if index == l {
		return
	}
	copy(e.buff[index+size:], e.buff[index:])
}

func (e *BlockWriter) grow(n int) int {
	l := len(e.buff)
	if cap(e.buff)-l > n {
		e.buff = e.buff[:l+n]
		return l
	}

	nb := make([]byte, l+n, cap(e.buff)*2+n)
	copy(nb, e.buff)
	e.buff = nb
	return l
}

// NewBlockReader returns a new BlockReader reading from r.
func NewBlockReader(r io.Reader) *BlockReader {
	return &BlockReader{
		buff: make([]byte, 0, 16),
		r:    r,
	}
}

// BlockReader provides a method for reading data in defined payloads.
type BlockReader struct {
	buff []byte
	off  int
	r    io.Reader
}

// Read implements io.Reader.
// If a previous call to Read has not completely read the payload it continues reading the previous payload,
// stopping at the end of the payload and returning EOF. If the previous payload has been read,
// it begins reading the next payload.
func (e *BlockReader) Read(buff []byte) (int, error) {
	if e.len() > 0 {
		c := copy(buff, e.buff[e.off:])
		e.off += c
		if e.len() == 0 {
			return c, io.EOF
		}
		return c, nil
	}

	n, err := e.getHeader()
	if err != nil {
		return 0, err
	}
	if n >= uint32(TooBig) {
		return 0, NewError(ErrMalformed, fmt.Sprintf("payload of size %v is too big", n), 0)
	}

	e.off = 0
	e.buff = e.buff[:0]
	e.grow(int(n))

	if err := Read(e.buff, e.r); err != nil {
		return 0, err
	}

	e.unescape(0, len(e.buff))

	c := copy(buff, e.buff)
	e.off += c
	if e.len() == 0 {
		return c, io.EOF
	}
	return c, nil
}

func (e *BlockReader) len() int { return len(e.buff) - e.off }

func (e *BlockReader) getHeader() (uint32, error) {
	e.buff = e.buff[:2]
	if err := Read(e.buff, e.r); err != nil {
		return 0, err
	}

	for e.buff[0] != EscapeByte || e.buff[1] == escapeEscape {
		e.buff[0] = e.buff[1]
		if err := Read(e.buff[1:], e.r); err != nil {
			return 0, err
		}
	}

	if e.buff[1] == EscapeByte {
		if err := Read(e.buff[1:2], e.r); err != nil {
			return 0, err
		}

		e.buff[1] = EscapeByte
	}

	n, size := DecodeVarUint32Header(e.buff[1])
	if size != 0 {
		l := e.grow(size)

		end := len(e.buff)
		for got := l; got < end; {
			if err := Read(e.buff[got:], e.r); err != nil {
				return 0, err
			}

			i := got
			got = end
			for ; i < got; i++ {
				if e.buff[i] == EscapeByte {
					end++
				}
			}

			e.grow(end - got)
		}

		e.unescape(l, end)

		n, _ = DecodeVarUint32(e.buff[1:])
	}

	return n, nil
}

func (e *BlockReader) unescape(x, y int) {
	for i := x; i < y; i++ {
		if e.buff[i] == EscapeByte {
			if i == y-1 {
				return
			}

			if e.buff[i+1] == escapeEscape {
				copy(e.buff[i+1:], e.buff[i+2:])
				e.buff = e.buff[:len(e.buff)-1]
				y--
				continue
			}
		}
	}
}

func (e *BlockReader) grow(n int) int {
	l := len(e.buff)
	if cap(e.buff)-l > n {
		e.buff = e.buff[:l+n]
		return l
	}

	nb := make([]byte, l+n, cap(e.buff)*2+n)
	copy(nb, e.buff)
	e.buff = nb
	return l
}
