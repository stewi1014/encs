package gram

import "io"

const (
	// MaxVarintBytes is the maximum write size of WriteUint and WriteInt.
	MaxVarintBytes = 9

	maxVarintByte = 256 - MaxVarintBytes // largest uint8 size for single byte transmission.
)

// WriteInt writes a variable-length encoding of n to the gram.
func (g *Gram) WriteInt(n int64) {
	enc := uint64(n) << 1
	if n < 0 {
		enc++
	}
	g.WriteUint(enc)
}

// WriteUint writes a variable-length encoding of n to the gram.
func (g *Gram) WriteUint(n uint64) {
	if n <= maxVarintByte {
		g.buff[g.grow(1)] = byte(n)
		return
	}

	size := 1
	for n > (1<<(8*size))-1 {
		size++
	}
	off := g.grow(size + 1)
	g.buff[off] = byte(maxVarintByte + size)
	for i := 1; n > 0; i++ {
		g.buff[off+i] = byte(n)
		n = n >> 8
	}
	return
}

// WriteUint32 writes a constant-length (4 byte) encoding of a uint32.
func (g *Gram) WriteUint32(n uint32) {
	buff := g.WriteBuff(4)
	buff[0] = uint8(n >> 24)
	buff[1] = uint8(n >> 16)
	buff[2] = uint8(n >> 8)
	buff[3] = uint8(n)
}

// Write implements io.Writer
func (g *Gram) Write(buff []byte) (int, error) {
	return copy(g.buff[g.grow(len(buff)):], buff), nil
}

// WriteByte implements io.ByteWriter
func (g *Gram) WriteByte(c byte) error {
	g.buff[g.grow(1)] = c
	return nil
}

// WriteBuff retrurns a slice for the next n bytes.
// The buffer must be written before other write calls.
func (g *Gram) WriteBuff(n int) []byte {
	return g.buff[g.grow(n):]
}

// LimitRead reads n bytes from r into the gram's buffer.
func (g *Gram) LimitRead(r io.Reader, n int) (err error) {
	if rg, ok := r.(*Gram); ok {
		if off, sliced := Sliced(rg.buff, g.buff); sliced {
			if len(g.buff)+off+n > len(rg.buff) {
				n = len(rg.buff) - off - len(g.buff)
				err = io.EOF
			}
			g.buff = rg.buff[off : off+len(g.buff)+n]
			return
		}
	}

	nb := g.buff[g.grow(n):]
	for len(nb) > 0 {
		c, err := r.Read(nb)
		nb = nb[c:]
		if err != nil {
			return err
		}
	}
	return nil
}

// WriteLater returns a gram for writing the next n bytes later.
// n bytes of data are written to the parent, and writes to the gram fill this space.
func (g *Gram) WriteLater(n int) *Gram {
	l := g.grow(n)
	return &Gram{
		buff:   SetCap(g.buff[l:l], n),
		parent: g,
		poff:   l,
	}
}

// Size returns the total size of the buffer.
func (g *Gram) Size() int {
	return len(g.buff)
}
