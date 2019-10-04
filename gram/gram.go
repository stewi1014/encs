package gram

import (
	"fmt"
	"io"
	"unsafe"
)

const (
	// 4 or 8 bytes for an int.
	wordSize = 4 << ((^uint(0) >> 32) & 1)

	// enables extra checks
	debug = true
)

// TooBig is a byte count used for simple sanity checking.
// By default it is 32MB on 32bit machines, and 128MB on 64bit machines.
// Feel free to change it.
var TooBig uint64 = 1 << (25 + ((^uint(0) >> 32) & 2))

func check(n uint64) {
	if n > TooBig {
		panic(fmt.Errorf("%v is too big", n))
	}
}

// NewGram returns a new gram.
func NewGram() *Gram {
	return &Gram{
		buff: getBuffer(),
	}
}

// ReadGram drains r, returning a Gram for reading the data.
func ReadGram(r io.Reader) (*Gram, error) {
	if g, ok := r.(*Gram); ok {
		return g, nil
	}

	g := NewGram()
	take := cap(g.buff)
read:
	l := g.grow(take)
	n, err := r.Read(g.buff[l:])
	if n == take && err == nil {
		take = take * 2
		goto read
	}
	g.buff = g.buff[:l+n]
	return g, err
}

// WriteGram returns a gram which will be written to w, and function to execute the write.
// It gives a speed boost if w is already a gram.
func WriteGram(w io.Writer) (g *Gram, write func() error) {
	if g, ok := w.(*Gram); ok {
		return g, func() error { return nil }
	}
	g = NewGram()
	return NewGram(), func() error {
		n, err := g.WriteTo(w)
		if err != nil {
			return err
		}
		if n != int64(g.Size()) {
			return fmt.Errorf("incomplete write; want %v bytes but only wrote %v", g.Size(), n)
		}
		g.Close()
		return nil
	}
}

// Gram provides methods for reading and writing buffers of data.
type Gram struct {
	buff []byte
	off  int

	parent *Gram
	poff   int
}

// Close releases buffers for later use.
func (g *Gram) Close() {
	if g.parent != nil {
		panic("close called on child gram")
	}
	putBuffer(g.buff)
}

func (g *Gram) grow(n int) (l int) {
	check(uint64(n))
	l = len(g.buff)
	c := cap(g.buff)
	if g.parent != nil {
		if c >= l+n {
			g.buff = g.parent.buff[g.poff : g.poff+l+n]
			return
		}
		panic(fmt.Errorf("cannot grow gram by %v (%v existing); is child gram and cap is %v", n, l, c))
	}
	if c >= l+n {
		g.buff = g.buff[:l+n]
		return
	}
	nb := make([]byte, l+n, c*2+n)
	copy(nb, g.buff)
	g.buff = nb
	return
}

func (g *Gram) setCap(n int) {
	if len(g.buff) > n {
		g.buff = g.buff[:n]
	}
	cptr := (uintptr)(unsafe.Pointer(&g.buff)) + (wordSize * 2)
	*(*int)(unsafe.Pointer(cptr)) = n
}
