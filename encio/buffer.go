package encio

import (
	"io"
	"sync"
)

// Buffer is a buffer for data. It operates similar to bytes.Buffer
type Buffer struct {
	buff []byte
	off  int
}

// Read implements io.Reader
func (b *Buffer) Read(buff []byte) (int, error) {
	n := copy(buff, b.buff[b.off:])
	b.off += n
	if n < len(buff) {
		return n, io.EOF
	}
	return n, nil
}

// ReadByte implements io.ByteReader
func (b *Buffer) ReadByte() (byte, error) {
	if b.Len() == 0 {
		return 0, io.EOF
	}
	by := b.buff[b.off]
	b.off++
	return by, nil
}

// Write implements io.Writer
func (b *Buffer) Write(buff []byte) (int, error) {
	return copy(b.buff[b.grow(len(buff)):], buff), nil
}

// WriteByte implements io.ByteWriter
func (b *Buffer) WriteByte(by byte) error {
	b.buff[b.grow(1)] = by
	return nil
}

// Len returns the length of the unread portion of the buffer
func (b *Buffer) Len() int {
	return len(b.buff) - b.off
}

func (b *Buffer) grow(n int) int {
	l := len(b.buff)
	if l+n <= cap(b.buff) {
		b.buff = b.buff[:l+n]
		return l
	}

	l -= b.off
	c := cap(b.buff)
	if (l+n)*8 <= c { // let cap grow to 8 time the size so we're not always sliding.
		// slide down
		copy(b.buff, b.buff[b.off:])
		b.buff = b.buff[:l+n]
		b.off = 0
		return l
	}
	// must allocate
	nb := make([]byte, l+n, c*2+n)
	copy(nb, b.buff[b.off:])
	b.buff = nb
	b.off = 0
	return l
}

// NewPipe creates a new Pipe
func NewPipe() *Pipe {
	return &Pipe{
		cond: sync.NewCond(new(sync.Mutex)),
	}
}

// Pipe is a buffered pipe. It operates like Buffer, but read calls will block until a call to write if the buffer is empty.
type Pipe struct {
	cond   *sync.Cond
	buff   []byte
	off    int
	closed bool
}

// Read implements io.Reader
func (p *Pipe) Read(buff []byte) (int, error) {
	p.cond.L.Lock()
	for p.len() == 0 && !p.closed {
		p.cond.Wait()
	}
	if p.closed {
		p.cond.L.Unlock()
		return 0, io.EOF
	}

	n := copy(buff, p.buff[p.off:])
	p.off += n
	p.cond.L.Unlock()
	return n, nil
}

// Write implements io.Writer
func (p *Pipe) Write(buff []byte) (int, error) {
	p.cond.L.Lock()
	if p.closed {
		p.cond.L.Unlock()
		return 0, io.ErrClosedPipe
	}

	n := copy(p.buff[p.grow(len(buff)):], buff)
	p.cond.Broadcast()

	p.cond.L.Unlock()
	return n, nil
}

// Close implements io.Closer
// Read calls will return io.EOF, and write calls will return io.ErrClosedPipe.
func (p *Pipe) Close() error {
	p.cond.L.Lock()
	p.closed = true
	p.cond.Broadcast()
	p.cond.L.Unlock()
	return nil
}

// len returns the length of the unread portion of the buffer
// mutex must be held
func (p *Pipe) len() int {
	return len(p.buff) - p.off
}

// mutex must be held
func (p *Pipe) grow(n int) int {
	l := len(p.buff)
	if l+n <= cap(p.buff) {
		p.buff = p.buff[:l+n]
		return l
	}

	l -= p.off
	c := cap(p.buff)
	if (l+n)*8 <= c { // let cap grow to 8 time the size so we're not always sliding.
		// slide down
		copy(p.buff, p.buff[p.off:])
		p.buff = p.buff[:l+n]
		p.off = 0
		return l
	}
	// must allocate
	nb := make([]byte, l+n, c*2+n)
	copy(nb, p.buff[p.off:])
	p.buff = nb
	p.off = 0
	return l
}
