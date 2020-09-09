package encio

import (
	"context"
	"errors"
	"fmt"
	"io"
	"sync"
	"time"
)

// This was quite difficult for me to write. In many places in encs I'm very happy to be able to logically split up code and write small,
// independent components that I'm confident in. In this case, I'm relying much more on testing (not that I don't already test!)
// rather than writing code I can look at and trust in the first place.
// I don't want the overhead of channels and a separate routine when the multiplexing functionality is not being used.
// I want to be able to instantiate a muxer on the off-chance that I actually end up needing to use it, and not have a noticeable performance impact if I don't.
// So, we have this annoyingly complex system where we switch between concurrent and non-concurrent modes depending on the existence of streams.
//
// I also doubt this performs well with massively concurrent use cases. That's not the point. Use https://github.com/xtaci/smux instead.
// The point is to provide the best single-stream/single-thread performance possible
// while providing multiplexing functionality *if* it's needed.

// The byte preceding the UUID for flags.
const (
	muxDefault = byte(1 << iota)
	muxStream
	muxData
	muxClose

	muxDefaultData = muxDefault | muxData
	muxStreamData  = muxStream | muxData
	muxStreamClose = muxStream | muxClose

	streamHeaderSize  = 17
	defaultheaderSize = 1
)

// NewMuxWriter returns a new MuxWriter writing to w.
func NewMuxWriter(w io.Writer) *MuxWriter {
	return &MuxWriter{
		w:       NewBlockWriter(w),
		streams: make(map[UUID]*muxWriterStream),
	}
}

// MuxWriter provides methods for multiplexing multiple streams onto a single stream.
//
// It is thread safe.
type MuxWriter struct {
	w       *BlockWriter
	mutex   sync.Mutex
	streams map[UUID]*muxWriterStream
}

// OpenStream creates a new stream, returning an io.WriteCloser for it.
// Closing the stream does not affect the MuxWriter.
func (m *MuxWriter) OpenStream(id UUID) (io.WriteCloser, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if m.w == nil {
		return nil, io.ErrClosedPipe
	}

	s := &muxWriterStream{
		id:  id,
		mux: m,
	}

	m.streams[id] = s
	return s, nil
}

// Write implements io.Writer.
// Write writes data on the default stream.
// If multiple MuxWriters are writing to the same buffer, data written here is merged.
func (m *MuxWriter) Write(buff []byte) (int, error) {
	if len(buff) == 0 {
		return 0, nil
	}

	m.mutex.Lock()
	defer m.mutex.Unlock()

	if m.w == nil {
		return 0, io.ErrClosedPipe
	}

	wbuff := m.w.Buff()
	wbuff.MWriteByte(muxDefaultData)
	wbuff.MWrite(buff)
	n, err := m.w.Write(*wbuff)
	n -= defaultheaderSize
	if n < 0 {
		n = 0
	}

	return n, err
}

// Close implements io.Closer.
// Close closes the MuxWriter.
// It is equivalent to calling Close() on all streams, but also prevents any further writes on the default stream.
// The remote MuxReader's default stream does not close.
func (m *MuxWriter) Close() error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if m.w == nil {
		return io.ErrClosedPipe
	}

	for id := range m.streams {
		if err := m.closeStream(id); err != nil {
			return err
		}
	}

	m.w = nil

	return nil
}

func (m *MuxWriter) closeStream(id UUID) error {
	if m.w == nil {
		return io.ErrClosedPipe
	}

	delete(m.streams, id)

	wbuff := m.w.Buff()
	wbuff.MWriteByte(muxStreamClose)
	wbuff.MWrite(id[:])

	n, err := m.w.Write(*wbuff)
	if n != streamHeaderSize && err == nil {
		err = io.ErrShortWrite
	}

	return err
}

type muxWriterStream struct {
	id  UUID
	mux *MuxWriter
}

func (m *muxWriterStream) Close() error {
	m.mux.mutex.Lock()
	defer m.mux.mutex.Unlock()

	return m.mux.closeStream(m.id)
}

func (m *muxWriterStream) Write(buff []byte) (int, error) {
	if len(buff) == 0 {
		return 0, nil
	}

	m.mux.mutex.Lock()
	defer m.mux.mutex.Unlock()

	if m.mux.w == nil {
		return 0, io.ErrClosedPipe
	}

	wbuff := m.mux.w.Buff()
	wbuff.MWriteByte(muxStreamData)
	wbuff.MWrite(m.id[:])
	wbuff.MWrite(buff)

	n, err := m.mux.w.Write(*wbuff)
	n -= streamHeaderSize
	if n < 0 {
		n = 0
	}

	return n, err
}

// NewMuxReader returns a new MuxReader.
func NewMuxReader(r io.Reader) *MuxReader {
	return &MuxReader{
		streams:         make(map[UUID]*muxStreamReader),
		r:               NewBlockReader(r),
		stream:          make(chan Buffer, 20),
		closed:          make(chan struct{}),
		unopenedStreams: make(map[UUID]*muxStreamReader),
	}
}

const readerTimeout = time.Second * 300

// MuxReader provides methods for reading multiplexed streams.
// It can begin reading from an arbitrary point in a stream, and can read from buffers that contain payloads from multiple MuxWriters
// in any order, so long as the payloads from each writing MuxWriter are in order with themselves.
//
// It silently ignores data from unopened streams, however it only discards unopened stream data that was received before the most recent default Read call.
type MuxReader struct {
	mutex           sync.Mutex
	unopenedStreams map[UUID]*muxStreamReader
	streams         map[UUID]*muxStreamReader
	r               *BlockReader
	concurrent      bool

	err    error
	closed chan struct{}
	stream chan Buffer
	buff   ReadBuffer
}

type muxStreamReader struct {
	mux *MuxReader

	closed chan struct{}
	stream chan Buffer
	id     UUID
	buff   ReadBuffer
}

func (m *muxStreamReader) Close() error {
	m.mux.mutex.Lock()
	defer m.mux.mutex.Unlock()

	_, ok := m.mux.streams[m.id]
	if ok {
		delete(m.mux.streams, m.id)
		close(m.closed)
		return nil
	}

	return io.ErrClosedPipe
}

func (m *muxStreamReader) Read(buff []byte) (int, error) {
	if n, err := m.readBuff(buff); err == nil {
		return n, nil
	}

	rbuff, ok := <-m.stream
	if !ok {
		m.mux = nil
		return 0, io.ErrClosedPipe
	}

	m.buff.Reset(rbuff)

	// Discard header bytes.
	if m.buff.Consume(streamHeaderSize) != nil {
		return 0, io.ErrUnexpectedEOF
	}

	return m.readBuff(buff)
}

func (m *muxStreamReader) readBuff(buff []byte) (int, error) {
	n, err := m.buff.Read(buff)
	if n > 0 && err == io.EOF {
		// Ignore error if it's the end of the buffer
		return n, nil
	}

	return n, err
}

// Open returns an io.ReadCloser for reading the given stream.
// It does not read anything from the underlying reader.
// The caller is responsible for resolving the ID with the sender.
func (m *MuxReader) Open(id UUID) (io.ReadCloser, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	concurrent, err := m.state()
	if err != nil {
		return nil, err
	}

	if !concurrent {
		m.runReads()
	}

	s, ok := m.unopenedStreams[id]
	if ok {
		delete(m.unopenedStreams, id)
	} else {
		s = m.newStream(id)
	}

	m.streams[id] = s
	return s, nil
}

func (m *MuxReader) newStream(id UUID) *muxStreamReader {
	return &muxStreamReader{
		mux:    m,
		stream: make(chan Buffer),
		id:     id,
		closed: make(chan struct{}),
	}
}

func (m *MuxReader) readConcurrent(buff []byte) (n int, ok bool, err error) {
	rbuff, ok := <-m.stream
	if !ok {
		return 0, false, nil
	}

	m.buff.Reset(rbuff)

	flag, err := m.buff.ReadByte()
	if err != nil {
		m.buff.Reset(nil)
		return 0, true, err
	}

	if flag&muxData|muxDefault != muxData|muxDefault {
		err = fmt.Errorf("unknown mux flag %v", flag)
	}

	n = m.buff.MRead(buff)
	return n, true, err
}

func (m *MuxReader) read(buff []byte) (int, error) {
readAgain:
	rbuff, err := m.r.ReadBytes()
	if err != nil {
		return 0, err
	}

	// Unfortunately we can't use the existing buffer outside the mutex.
	// We could transition to concurrent mode, and the buffer read to before our client has consumed all the data,
	// overwriting it.
	m.buff.MWrite(*rbuff)

	flag, err := m.buff.ReadByte()
	if err != nil {
		return 0, err
	}

	if flag&muxData|muxDefault != muxData|muxDefault {
		// Discard the payload.
		goto readAgain
	}

	n := m.buff.MRead(buff)
	if n > 0 {
		return n, nil
	}
	return 0, NewIOError(
		ErrMalformed,
		m.r,
		"sent payload is empty",
		0,
	)
}

func (m *MuxReader) Read(buff []byte) (int, error) {
readAgain:
	n := m.buff.MRead(buff)
	if n > 0 {
		return n, nil
	}

	m.mutex.Lock()
	concurrent, err := m.state()
	if err != nil {
		m.mutex.Unlock()
		return 0, err
	}

	if concurrent {
		m.mutex.Unlock()
		n, ok, err := m.readConcurrent(buff)
		if ok {
			return n, err
		}
		goto readAgain
	}

	defer m.mutex.Unlock()
	return m.read(buff)
}

func (m *MuxReader) runReads() {
	c, err := m.state()
	if err != nil {
		panic(err)
	}

	if c {
		panic("reader already running")
	}

	m.concurrent = true
	m.stream = make(chan Buffer)

	go func(r *BlockReader) {
		timer := time.NewTimer(readerTimeout)

		fail := func(err error) {
			if cerr := m.close(); cerr != nil {
				m.err = cerr
			} else {
				m.err = err
			}
			close(m.stream)
		}

		writeChan := func(buff Buffer, stream chan Buffer, closed chan struct{}) error {
			timer.Reset(readerTimeout)
			select {
			case stream <- buff:
				return nil
			case <-closed:
				return io.ErrClosedPipe
			case <-timer.C:
				fmt.Fprintf(Warnings, "Reader timed out. MuxReader is not a buffer. If you're not reading from it either close it or don't stop reading from open streams.\n")
				return context.DeadlineExceeded
			}
		}

		m.mutex.Lock()

		for {
			// Check if we can exit concurrent mode
			if len(m.streams) == 0 {
				m.concurrent = false
				m.mutex.Unlock()
				close(m.stream)
				return
			}

			m.mutex.Unlock()
			buff, err := r.ReadBytes()
			m.mutex.Lock()

			r = m.r
			if r == nil {
				m.mutex.Unlock()
				return
			}

			if err != nil {
				if errors.Is(err, ErrMalformed) {
					continue
				}
				fail(err)
				m.mutex.Unlock()
				return
			}

			if buff.Len() <= defaultheaderSize {
				fmt.Fprintf(Warnings, "Ignoring payload of %v bytes, smaller than the header size of %v bytes\n", buff.Len(), defaultheaderSize)
				continue
			}

			flag := (*buff)[0]

			switch {
			case flag&muxDefaultData == muxDefaultData:
				var nb Buffer
				nb.MWrite(*buff)

				m.mutex.Unlock()
				err := writeChan(nb, m.stream, m.closed)
				m.mutex.Lock()

				if err != nil {
					fail(err)
					return
				}

			case flag&muxStreamData == muxStreamData:
				if buff.Len() < streamHeaderSize {
					fmt.Fprintf(Warnings, "Ignoring payload of %v bytes, smaller than the header size of %v bytes\n", buff.Len(), streamHeaderSize)
					break
				}

				var id UUID
				copy(id[:], (*buff)[1:17])

				var nb Buffer
				s, ok := m.streams[id]
				us, uok := m.unopenedStreams[id]

				nb.MWrite(*buff)

				switch {
				case ok:
					m.mutex.Unlock()
					err := writeChan(nb, s.stream, s.closed)
					if err != nil {
						close(s.stream)
					}
					m.mutex.Lock()
				case uok:
					us.buff.MWrite(nb[streamHeaderSize:])
				default:
					s = m.newStream(id)
					s.buff.MWrite(nb[streamHeaderSize:])
					m.unopenedStreams[id] = s
				}
			case flag&muxStreamClose == muxStreamClose:
				if buff.Len() < streamHeaderSize {
					fmt.Fprintf(Warnings, "Ignoring payload of %v bytes, smaller than the header size of %v bytes\n", buff.Len(), streamHeaderSize)
					break
				}

				var id UUID
				copy(id[:], (*buff)[1:17])

				s, ok := m.streams[id]
				if ok {
					delete(m.streams, id)
					close(s.stream)
				}
			default:
				fmt.Fprintf(Warnings, "Ignoring packet with unknown header %08b\n", flag)
			}
		}
	}(m.r)
}

func (m *MuxReader) state() (concurrent bool, err error) {
	err = m.err
	if err == nil && m.r == nil {
		err = io.ErrClosedPipe
	}
	concurrent = m.concurrent || len(m.stream) > 0
	return
}

// Close implements io.Closer.
// It prevents any further read calls to the underlying reader.
func (m *MuxReader) Close() error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	return m.close()
}

func (m *MuxReader) close() error {
	if m.r == nil {
		// already closed
		return m.err
	}

	for _, s := range m.streams {
		close(s.stream)
	}

	close(m.closed)

	m.r = nil

	return m.err
}
