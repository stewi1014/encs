package gram

import (
	"fmt"
	"io"
)

// ReadInt reads a variable-length encoding of an int from the gram.
func (g *Gram) ReadInt() int64 {
	enc := g.ReadUint()
	sb := enc & 1
	enc = (enc >> 1) | (sb << 63)
	return int64(enc)
}

// ReadUint reads a variable-length encoding of a uint from the gram.
func (g *Gram) ReadUint() (n uint64) {
	b := g.buff[g.off]
	g.off++
	if b <= maxVarintByte {
		n = uint64(b)
		return
	}

	size := int(b - maxVarintByte)
	if size > g.Len() {
		panic("not enough data to read varint")
	}
	for i := 0; i < size; i++ {
		n += uint64(g.buff[g.off+i]) << (8 * i)
	}
	g.off += size
	return
}

// ReadUint32 reads a constant-length (4 byte) encoding of a uint32
func (g *Gram) ReadUint32() (n uint32) {
	buff := g.ReadBuff(4)
	n += uint32(buff[0]) << 24
	n += uint32(buff[1]) << 16
	n += uint32(buff[2]) << 8
	n += uint32(buff[3])
	return
}

// ReadBuff reads n bytes, returning a slice for the read region.
func (g *Gram) ReadBuff(n int) []byte {
	if g.Len() < n {
		n = g.Len()
	}
	g.off += n
	return g.buff[g.off-n : g.off]
}

// Read implements io.Reader
func (g *Gram) Read(buff []byte) (int, error) {
	c := copy(buff, g.buff[g.off:])
	g.off += c
	if g.off == len(g.buff) {
		return c, io.EOF
	}
	return c, nil
}

// ReadAll returns the unread portion of the buffer.
func (g *Gram) ReadAll() []byte {
	buff := g.buff[g.off:]
	g.off = len(g.buff)
	return buff
}

// ReadByte implements io.ByteReader
func (g *Gram) ReadByte() (byte, error) {
	if g.off == len(g.buff) {
		return 0, io.EOF
	}
	g.off++
	return g.buff[g.off-1], nil
}

// LimitReader returns a reader (*Gram) for reading the next n bytes,
// returning io.EOF after n bytes read.
// If there is less than n bytes remaining, only the remaining bytes are read.
func (g *Gram) LimitReader(n int) *Gram {
	check(uint64(n))
	if g.Len() < n {
		n = g.Len()
	}
	g.off += n
	return &Gram{
		buff: SetCap(g.buff[g.off-n:g.off], n),
	}
}

// WriteTo implements io.WriterTo
func (g *Gram) WriteTo(w io.Writer) (int64, error) {
	l := g.Len()
	n, err := w.Write(g.buff[g.off:])
	g.off += n

	if err != nil {
		return int64(n), err
	}

	if n != l {
		return int64(n), fmt.Errorf("short write, want %v got %v", l, n)
	}

	return int64(n), err
}

// Len returns the size of the unread portion of the buffer.
func (g *Gram) Len() int {
	return len(g.buff) - g.off
}
