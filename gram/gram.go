package gram

import (
	"fmt"
	"io"
	"unsafe"
)

const (
	// TooBig is a byte count used for simple sanity checking.
	// By default it is 32MB on 32bit machines, and 128MB on 64bit machines.
	// Feel free to change it.
	TooBig = 1 << (25 + ((^uint(0) >> 32) & 2))

	// 4 or 8 bytes for an int.
	wordSize = 4 << ((^uint(0) >> 32) & 1)

	// enables extra checks
	debug = true
)

var (
	// NewGramSize is the minumum buffer size in bytes that a new Gram will have.
	NewGramSize = 32
)

func check(n uint64) {
	if n > TooBig {
		panic(fmt.Errorf("%v is too big", n))
	}
}

// NewGram returns a new gram.
func NewGram() *Gram {
	return &Gram{
		buff: GetBuffer(NewGramSize),
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

// ReadNGram is the same as ReadGram, only it stops at n bytes
func ReadNGram(r io.Reader, n int) (g *Gram, err error) {
	if g, ok := r.(*Gram); ok {
		return g.LimitReader(n), nil
	}

	g = NewGram()
	l := g.grow(n)
	var read int
	for l < n {
		read, err = r.Read(g.buff[l:n])
		if err != nil {
			break
		}
		if n == 0 {
			err = io.EOF
			break
		}
		l += read
	}
	g.buff = g.buff[:l]
	return
}

// WriteGram returns a gram which will be written to w, and function to execute the write.
// It returns r if r is already a Gram, and returns an error on short writes.
func WriteGram(w io.Writer) (g *Gram, write func() error) {
	if g, ok := w.(*Gram); ok {
		return g, func() error { return nil }
	}
	g = NewGram()
	return g, func() error {
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

type buffer struct {
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
	PutBuffer(g.buff)
}

// Reset clears the Gram's buffer, retaining space for later use.
func (g *Gram) Reset() {
	g.buff = g.buff[:0]
	g.off = 0
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

// Slide buffer abover index by slide up
func (g *Gram) slide(index, slide int) {
	l := len(g.buff)
	if index >= l {
		panic(fmt.Errorf("index %v out of bounds; len %v", index, l))
	}
	if slide >= 0 {
		g.grow(slide)
		copy(g.buff[index+slide:], g.buff[index:])
		return
	}
	if index-slide >= l {
		// Nothing to copy; trim buffer and return
		g.buff = g.buff[:index]
		return
	}
	copy(g.buff[index:], g.buff[index-slide:])
	g.buff = g.buff[:l+slide]
	return
}

func (g *Gram) setCap(n int) {
	if len(g.buff) > n {
		g.buff = g.buff[:n]
	}
	cptr := (uintptr)(unsafe.Pointer(&g.buff)) + (wordSize * 2)
	*(*int)(unsafe.Pointer(cptr)) = n
}
