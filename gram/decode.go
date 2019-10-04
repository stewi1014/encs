package gram

import "io"

// Reader is a method for reading grams.
type Reader interface {
	Read() (g *Gram, err error)
}

func NewStreamReader(r io.Reader) StreamReader {
	return StreamReader{
		r: r,
	}
}

type StreamReader struct {
	r io.Reader
}

func (sr StreamReader) Read() (*Gram, error) {
	g := NewGram()
	err := g.LimitRead(sr.r, 4)
	if err != nil {
		return g, err
	}
	l := g.ReadUint32()
	check(uint64(l))
	return g, g.LimitRead(sr.r, int(l))
}
