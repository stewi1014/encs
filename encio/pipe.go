package encio

import (
	"io"
	"sync"
)

// Pipe returns a synchronous, buffered pipe.
// A call to close results in subsequent calls to Write returning io.ErrClosedPipe,
// while read will continue reading the buffer before returning io.ErrClosedPipe when it is empty.
// It uses mutexes.
func Pipe() (*PipeReader, *PipeWriter) {

	cond := sync.NewCond(new(sync.Mutex))
	buff := new(ReadBuffer)
	err := new(error)

	return &PipeReader{
			cond: cond,
			buff: buff,
			err:  err,
		}, &PipeWriter{
			cond: cond,
			buff: buff,
			err:  err,
		}
}

// PipeReader implements the reading half of a pipe.
type PipeReader struct {
	cond *sync.Cond
	buff *ReadBuffer
	err  *error
}

func (p *PipeReader) Read(buff []byte) (n int, err error) {
	p.cond.L.Lock()
	defer p.cond.L.Unlock()

	n, err = p.readBuff(buff)
	if n > 0 || err != nil {
		return
	}

	for p.buff.Len() <= 0 && *p.err == nil {
		p.cond.Wait()
	}

	n, err = p.readBuff(buff)
	return
}

func (p *PipeReader) Close() error {
	p.cond.L.Lock()
	defer p.cond.L.Unlock()
	if *p.err != nil {
		return *p.err
	}

	p.cond.Broadcast()
	*p.err = io.ErrClosedPipe
	return nil
}

func (p *PipeReader) readBuff(buff []byte) (int, error) {
	if p.buff.Len() > 0 {
		n, _ := p.buff.Read(buff)
		if n < len(buff) {
			return n, *p.err
		}

		return n, nil
	}

	return 0, *p.err
}

// PipeWriter implements the writing half of a pipe.
type PipeWriter struct {
	cond *sync.Cond
	buff *ReadBuffer
	err  *error
}

func (p *PipeWriter) Write(buff []byte) (n int, err error) {
	p.cond.L.Lock()
	defer p.cond.L.Unlock()
	defer p.cond.Broadcast()

	if *p.err != nil {
		err = *p.err
		return
	}

	n, err = p.buff.Write(buff)
	return
}

func (p *PipeWriter) Close() error {
	p.CloseWith(io.ErrClosedPipe)
	return nil
}

func (p *PipeWriter) CloseWith(err error) {
	p.cond.L.Lock()
	defer p.cond.L.Unlock()
	p.cond.Broadcast()
	*p.err = err
}
