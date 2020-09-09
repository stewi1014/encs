package encio

import (
	"errors"
	"fmt"
	"hash"
	"hash/crc32"
	"io"
)

// NewChecksumWriter returns a new ChecksumWriter reading from r, using the given hasher for checking consistency.
// The ChecksumWriter and ChecksumReader must share the same hasher.
func NewChecksumWriter(w io.Writer, hasher hash.Hash) *ChecksumWriter {
	if hasher == nil {
		hasher = crc32.New(crc32.IEEETable)
	}

	return &ChecksumWriter{
		w:      w,
		hasher: hasher,
		header: make([]byte, hasher.Size()+4),
	}
}

// ChecksumWriter provides methods for detecting errors in data transmission.
type ChecksumWriter struct {
	w      io.Writer
	hasher hash.Hash
	header Buffer
	count  uint16
}

// Count returns the count of the next packet to be sent.
func (c *ChecksumWriter) Count() uint16 {
	return c.count
}

// Write implements io.Writer.
// It adds a checksum and incrementing integer to written data,
// allowing a subsequent ChecksumReader to discern if data is intact and in order.
func (c *ChecksumWriter) Write(buff []byte) (int, error) {
	if len(buff) == 0 {
		return 0, nil
	}

	hs := c.hasher.Size()

	c.header = c.header[:0]
	c.header.Grow(hs + 6)
	EncodeUint32(c.header[hs:], uint32(len(buff)))

	c.header[hs+4] = byte(c.count)
	c.header[hs+5] = byte(c.count >> 8)
	c.count++

	err := c.writeChecksum(c.header[hs:], buff)
	if err != nil {
		return 0, err
	}

	err = Write(c.header, c.w)
	if err != nil {
		return 0, err
	}

	return c.w.Write(buff)
}

func (c *ChecksumWriter) writeChecksum(buffs ...[]byte) error {
	c.hasher.Reset()

	for _, buff := range buffs {
		n, err := c.hasher.Write(buff)
		if n != len(buff) || err != nil {
			if err == nil {
				err = NewError(io.ErrShortWrite, fmt.Sprintf("hasher reported %v bytes written when given %v bytes", n, len(buff)), 0)
			}
			return err
		}
	}

	copy(c.header, c.hasher.Sum(c.header[:0]))

	return nil
}

// NewChecksumReader returns a ChecksumReader writing from w, using the given hasher for checking consistency.
func NewChecksumReader(r io.Reader, hasher hash.Hash) *ChecksumReader {
	if hasher == nil {
		hasher = crc32.New(crc32.IEEETable)
	}

	return &ChecksumReader{
		r:      r,
		hasher: hasher,
	}
}

// ChecksumReader provides methods for detecting errors in data transmission.
type ChecksumReader struct {
	r        io.Reader
	hasher   hash.Hash
	buff     Buffer
	off      int
	count    uint16
	hasCount bool
}

func (c *ChecksumReader) reset() {
	c.buff = c.buff[:0]
	c.hasCount = false
	c.off = 0
}

// Count returns the next expected packet number.
func (c *ChecksumReader) Count() uint16 {
	return c.count
}

// Read implements io.Reader.
// If received data is out of order, or if data has been corrupted as discerned by the hasher,
// it will return a wrapped ErrMalformed.
//
// If the reader returns io.ErrUnexpectedEOF or io.EOF, it replaces the error with ErrMalformed if the header's reported size doesn't match the number of received bytes.
// A call after the previous call returned an error will ignore the next received packet number, and continue checking order on subsequent calls.
func (c *ChecksumReader) Read(buff []byte) (int, error) {
	if len(c.buff)-c.off > 0 {
		n := copy(buff, c.buff[c.off:])
		c.off += n
		return n, nil
	}

	if len(buff) == 0 {
		return 0, nil
	}

	return c.read(buff)
}

func (c *ChecksumReader) read(buff []byte) (int, error) {
	hs := c.hasher.Size()

	l, herr := c.readHeader()
	berr := c.readBody(l)
	if herr != nil {
		c.reset()
		return 0, herr
	}
	if berr != nil {
		c.reset()
		return 0, berr
	}

	if err := c.check(); err != nil {
		c.reset()
		return 0, err
	}

	c.off = hs + 6
	n := copy(buff, c.buff[c.off:])
	c.off += n
	return n, nil
}

func (c *ChecksumReader) check() error {
	hs := c.hasher.Size()

	c.hasher.Reset()
	_, err := c.hasher.Write(c.buff[hs:])
	if err != nil {
		c.reset()
		return err
	}

	sum := c.hasher.Sum(make([]byte, 0, hs))

	for i := 0; i < len(sum); i++ {
		if c.buff[i] != sum[i] {
			c.reset()
			return NewIOError(
				ErrMalformed,
				c.r,
				"checksums do not match",
				0,
			)
		}
	}

	return nil
}

func (c *ChecksumReader) readHeader() (int, error) {
	hs := c.hasher.Size()

	c.buff = c.buff[:0]
	c.buff.Grow(hs + 6)

	if err := Read(c.buff, c.r); err != nil {
		return 0, err
	}

	l := int(DecodeUint32(c.buff[hs:]))
	if l > int(TooBig) {
		return 0, NewIOError(
			ErrMalformed,
			c.r,
			fmt.Sprintf(
				"received block size of %v is too big!",
				l,
			),
			0,
		)
	}

	count := uint16(c.buff[hs+4])
	count |= uint16(c.buff[hs+5]) << 8

	if count != c.count && c.hasCount {
		return l, NewIOError(
			ErrMalformed,
			c.r,
			fmt.Sprintf(
				"data out of order. Last packet was number %v, but just received %v",
				c.count,
				count,
			),
			0,
		)
	}
	c.count = count + 1
	c.hasCount = true

	return l, nil
}

func (c *ChecksumReader) readBody(n int) error {
	hs := c.headerSize()

	c.buff.Grow(n)

	if err := Read(c.buff[hs:], c.r); err != nil {
		c.reset()
		if errors.Is(err, io.ErrUnexpectedEOF) || errors.Is(err, io.EOF) {
			return NewIOError(
				ErrMalformed,
				c.r,
				fmt.Sprintf("header says the payload is %v bytes big, but got \"%v\" before reading it all", n, err),
				0,
			)
		}
		return err
	}

	return nil
}

func (c *ChecksumReader) headerSize() int {
	return c.hasher.Size() + 4 + 2
}
