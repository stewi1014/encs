package gram

import (
	"fmt"
	"io"
)

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

func NewPromiscuousReader(r io.Reader) PromiscuousReader {
	return PromiscuousReader{
		r: r,
		b: make([]byte, 1),
	}
}

type PromiscuousReader struct {
	r io.Reader
	b []byte
}

func (pr PromiscuousReader) Read() (*Gram, error) {
	// find the next instances of escapeByte
findEscapeByte:
	for {
		n, err := pr.r.Read(pr.b)
		if n != 1 {
			if err != nil {
				return nil, err
			}
			return nil, fmt.Errorf("incomplete read")
		}
		if pr.b[0] == escapeByte {
			break
		}
	}
	// we have an escape byte, check it's followed by a gramStart byte. If not, resume searching.

	n, err := pr.r.Read(pr.b)
	if n != 1 {
		if err != nil {
			return nil, err
		}
		return nil, fmt.Errorf("incomplete read")
	}
	if pr.b[0] != gramBegin {
		// Not a gram start, continute searching
		goto findEscapeByte
	}

	// we have a gram start header.
	g := NewGram()

	// read length after un-escaping
	err = g.LimitRead(pr.r, 4)
	if err != nil {
		return nil, err
	}

	i := new(Iter)
	for i.IterTo(g, 4) {
		if g.Byte(i) == escapeByte {
			err := g.LimitRead(pr.r, 1)
			if err != nil {
				return nil, err
			}
			i.Delta(1)
			if g.Byte(i) == escapeEscape {
				g.Remove(i)
			} else {
				panic("non-escaped escape byte in message")
			}
		}
	}

	// we've read 4-unescaped bytes, read l bytes, escape and return.

	l := g.ReadUint32()
	check(uint64(l))
	err = g.LimitRead(pr.r, int(l))
	if err != nil {
		return nil, err
	}

	for i.Iter(g) {
		if g.Byte(i) == escapeByte {
			if !i.Iter(g) {
				panic("escape byte at end of buffer")
			}
			if g.Byte(i) == escapeEscape {
				g.Remove(i)
			} else {
				panic("non-escaped escape byte in message")
			}
		}
	}

	return g, nil
}
