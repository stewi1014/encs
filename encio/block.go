package encio

import (
	"fmt"
	"io"
)

const (
	// escapeByte is the byte used to define the start of a block.
	// It should be reasonably uncommon.
	escapeByte byte = 23

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
// Headers can be placed at the start of the payload passed to write,
// and they will always be at the beginning of the payload when read using BlockReader.
// It is stream-promiscuous; any number of BlockWriters can share the same buffer,
// or change buffers mid-stream, and be read successfully.
type BlockWriter struct {
	buff []byte
	w    io.Writer
}

// Write implements io.Writer.
// The buffer passed to Write is immediately written along with a 5 byte header in a single call to Write on the wrapped reader.
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
	e.buff[4-n] = escapeByte

	e.escape(5-n, 5)

	return c, Write(e.buff[4-n:], e.w)
}

func (e *BlockWriter) escape(x, y int) {
	for i := x; i < y; i++ {
		if e.buff[i] == escapeByte {
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
// It reads up to the end of the payload that was passed to a call to Write, returns io.EOF, then continues reading the next payload on the next call.
// Typically, this is undesierable behaviour for a stream, as the assumption that when io.EOF is returned the Reader has no more data is false.
// Callers should be aware of this when using BlockReader.
//
// This does, however, allow some interesting things. BlockReader is buffer-promiscuous.
// In the extreme, this means that any number of BlockWriters can be writing to a single buffer in any order,
// and any number of BlockReaders can begin reading at any point in the buffer while only returning fullly-formed payloads.
// Any payload that BlockReader begins reading halfway through will be discarded until the next complete payload.
type BlockReader struct {
	buff []byte
	off  int
	r    io.Reader
}

// Read implements io.Reader.
//
// If a previous call to Read has not completely read the payload it continues reading the previous payload,
// stopping at the end of the payload and returning io.EOF. If the previous payload has been read,
// it begins reading the next payload.
//
// If the underlying reader returns io.EOF, read will return 0, io.EOF when the end of the current payload is reached.
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
		return 0, NewIOError(ErrTooBig, e.r, fmt.Sprintf("payload of size %v", n), 0)
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
	n, err := e.r.Read(e.buff)
	if err != nil {
		return 0, err
	}
	if n != 2 {
		return 0, io.ErrNoProgress
	}

	for e.buff[0] != escapeByte || e.buff[1] == escapeEscape {
		e.buff[0] = e.buff[1]
		n, err = e.r.Read(e.buff[1:])
		if err != nil {
			return 0, err
		}
		if n != 1 {
			return 0, io.ErrNoProgress
		}
	}

	if e.buff[1] == escapeByte {
		n, err = e.r.Read(e.buff[1:2])
		if err != nil {
			return 0, err
		}
		if n != 1 {
			return 0, io.ErrNoProgress
		}

		e.buff[1] = escapeByte
	}

	h, size := DecodeVarUint32Header(e.buff[1])
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
				if e.buff[i] == escapeByte {
					end++
				}
			}

			e.grow(end - got)
		}

		e.unescape(l, end)

		h, _ = DecodeVarUint32(e.buff[1:])
	}

	return h, nil
}

func (e *BlockReader) unescape(x, y int) {
	for i := x; i < y; i++ {
		if e.buff[i] == escapeByte {
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
