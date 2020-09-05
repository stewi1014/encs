package encio

import (
	"io"
	"sync"
)

// Pipe returns a synchronous, buffered pipe.
// A call to close results in subsequent calls to Write returning io.ErrClosedPipe,
// while read will continue reading until the buffer until returning io.ErrClosedPipe.
// It simply uses mutexes.
func Pipe() io.ReadWriteCloser {
	return &pipe{
		cond: sync.NewCond(&sync.Mutex{}),
		buff: make([]byte, 0, 256),
	}
}

type pipe struct {
	cond   *sync.Cond
	buff   []byte
	off    int
	closed bool
}

func (p *pipe) Read(buff []byte) (int, error) {
	p.cond.L.Lock()
	defer p.cond.L.Unlock()

	if len(p.buff)-p.off != 0 {
		n := copy(buff, p.buff[p.off:])
		p.off += n
		return n, nil
	}

	if p.closed {
		return 0, io.ErrClosedPipe
	}

	for len(p.buff)-p.off <= 0 && !p.closed {
		p.cond.Wait()
	}

	n := copy(buff, p.buff[p.off:])
	p.off += n

	if p.closed && n < len(buff) {
		return n, io.ErrClosedPipe
	}

	return n, nil
}

func (p *pipe) Write(buff []byte) (int, error) {
	p.cond.L.Lock()
	defer p.cond.L.Unlock()
	defer p.cond.Broadcast()

	if p.closed {
		return 0, io.ErrClosedPipe
	}

	l := len(p.buff)
	wl := len(buff)
	if cap(p.buff) >= l+wl {
		// Have enough space
		p.buff = p.buff[:l+wl]
		wl = copy(p.buff[l:], buff)
		return wl, nil
	}

	if cap(p.buff) >= ((l+wl)-p.off)*8 {
		// Slide, but only if we can reclaim a reasonably large amount of data.
		copy(p.buff, p.buff[p.off:])
		wl = copy(p.buff[l-p.off:], buff)
		p.buff = p.buff[:l+wl-p.off]
		p.off = 0
		return wl, nil
	}

	// Allocate
	nb := make([]byte, (l-p.off)+wl, cap(p.buff)+wl)
	copy(nb, p.buff[p.off:])
	wl = copy(nb[l-p.off:], buff)
	p.buff = nb
	p.off = 0
	return wl, nil
}

func (p *pipe) Close() error {
	p.cond.L.Lock()
	defer p.cond.L.Unlock()
	p.closed = true
	return nil
}
