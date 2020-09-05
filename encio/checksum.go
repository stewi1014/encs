package encio

import (
	"errors"
	"fmt"
	"hash"
	"hash/crc32"
	"io"
)

// NewChecksumWriter returns a new ChecksumWriter using the given hasher, and writing to w.
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

// BlockSize returns the hash's underlying block size.
// The Write method can accept any amount
// of data, but it may operate more efficiently if all writes
// are a multiple of the block size depending on the hasher implementation.
func (c *ChecksumWriter) BlockSize() int { return c.hasher.BlockSize() }

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

	c.hasher.Reset()
	_, err := c.hasher.Write(c.header[hs:])
	if err != nil {
		return 0, err
	}

	_, err = c.hasher.Write(buff)
	if err != nil {
		return 0, err
	}

	copy(c.header, c.hasher.Sum(c.header[:0]))

	c.count++

	if err := Write(c.header, c.w); err != nil {
		return 0, err
	}

	return c.w.Write(buff)
}

// NewChecksumReader returns a ChecksumReader using the given hasher,
// and reading from r.
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
	r      io.Reader
	hasher hash.Hash
	buff   Buffer
	off    int
	count  int
}

func (c *ChecksumReader) reset() {
	c.buff = c.buff[:0]
	c.count = 0
	c.off = 0
}

// Read implements io.Reader.
// If received data is out of order, or if data has been corrupted as discerned by the hashing algorithm,
// Read will return a wrapped ErrMalformed.
//
// If the reader returns io.ErrUnexpectedEOF or io.EOF, Read replaces the error with ErrMalformed if the header's reported size doesn't match the number of received bytes.
// A call to Read after the last call returned an error will ignore the received packet number, and continue checking order on subsequent calls.
func (c *ChecksumReader) Read(buff []byte) (int, error) {
	if len(c.buff)-c.off > 0 {
		n := copy(buff, c.buff[c.off:])
		c.off += n
		return n, nil
	}

	if len(buff) == 0 {
		return 0, nil
	}

	hs := c.hasher.Size()

	c.buff = c.buff[:0]
	c.buff.Grow(hs + 6)

	if err := Read(c.buff, c.r); err != nil {
		c.reset()
		return 0, err
	}

	l := int(DecodeUint32(c.buff[hs:]))
	if l > int(TooBig) {
		c.reset()
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

	if int(count) != c.count && c.count >= 0 {
		last := c.count
		received := count
		c.reset()
		return 0, NewIOError(
			ErrMalformed,
			c.r,
			fmt.Sprintf(
				"data out of order. Last packet was number %v, but just received %v",
				last,
				received,
			),
			0,
		)
	}
	c.count = int(count + 1)

	c.buff.Grow(l)

	if err := Read(c.buff[hs+6:], c.r); err != nil {
		c.reset()
		if errors.Is(err, io.ErrUnexpectedEOF) || errors.Is(err, io.EOF) {
			return 0, NewIOError(
				ErrMalformed,
				c.r,
				fmt.Sprintf("header says the payload is %v bytes big, but got \"%v\" before reading it all.", l, err),
				0,
			)
		}
		return 0, err
	}

	c.hasher.Reset()
	_, err := c.hasher.Write(c.buff[hs:])
	if err != nil {
		c.reset()
		return 0, err
	}

	sum := c.hasher.Sum(make([]byte, 0, hs))

	for i := 0; i < len(sum); i++ {
		if c.buff[i] != sum[i] {
			c.reset()
			return 0, NewIOError(
				ErrMalformed,
				c.r,
				"checksums do not match",
				0,
			)
		}
	}

	c.off = hs + 6
	n := copy(buff, c.buff[c.off:])
	c.off += n
	return n, nil
}
