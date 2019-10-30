// Package encio provides simple methods for encoding-relavent input and output, as well as error types.
// It is split from encodable to avoid its cluttered namespace.
package encio

import (
	"errors"
	"fmt"
	"io"
)

var (
	// TooBig is a byte count used for simple sanity checking before things like allocation and iteration with numbers decoded from readers.
	// ErrMalformed is returned if a metric exceeds this.
	//
	// By default it is 32MB on 32bit machines, and 128MB on 64bit machines.
	// Feel free to change it.
	TooBig = 1 << (25 + ((^uint(0) >> 32) & 2))
)

// Read reads from r, filling buff and handling errors of io.Reader with little overhead in almost all cases.
// In an ideaĺ read, only a single int equality check is performed. If the read reports the whole buffer is written, returned errors are ignored.
func Read(buff []byte, r io.Reader) error {
	n, err := r.Read(buff)
	if n == len(buff) {
		return nil
	}

	end := n
	for end < len(buff) && err == nil && n > 0 {
		n, err = r.Read(buff[end:])
		end += n
	}

	if end != len(buff) {
		switch {
		case end > len(buff):
			return IOError{
				Err:     errors.New("bad io.Reader implementation"),
				Message: fmt.Sprintf("Read() reported %v bytes read, but buffer is only %v bytes", end, len(buff)),
			}
		case err == io.EOF:
			return IOError{
				Err:     err,
				Message: fmt.Sprintf("unexpected EOF, want %v bytes but only got %v", len(buff), end),
			}
		case err != nil:
			return IOError{
				Err:     err,
				Message: fmt.Sprintf("want %v bytes but only got %v bytes", len(buff), end),
			}
		default: //err == nil
			return IOError{
				Err:     io.ErrNoProgress,
				Message: fmt.Sprintf("read %v bytes, need %v bytes", end, len(buff)),
			}
		}
	}
	return nil
}

// Write writes to w from buff, handling errors of io.Writer with little overhead in almost all cases.
// In an ideal write, only a single int equality check is performed. If the write reports the whole buffer is written, returned errors are ignored.
func Write(buff []byte, w io.Writer) error {
	n, err := w.Write(buff)
	if n == len(buff) {
		return nil
	}

	end := n
	for end < len(buff) && err == nil && n > 0 {
		// This isn't quite to spec; multiple writes shouldn't be neccecary, but I don't see the harm in this.
		// This loop will only run if n from the first write is >0 and <len(buff), and err == nil.
		// That seems like a pretty safe case to try writing more data, and introduces no overhead to normal writes.
		n, err = w.Write(buff[end:])
		end += n
	}

	if end != len(buff) {
		switch {
		case end > len(buff):
			return IOError{
				Err:     errors.New("bad io.Writer implementation"),
				Message: fmt.Sprintf("Write() reported %v bytes written, but buffer is only %v bytes", end, len(buff)),
			}
		case err == nil:
			return IOError{
				Err:     io.ErrShortWrite,
				Message: fmt.Sprintf("want %v bytes but only wrote %v bytes with no error", len(buff), end),
			}
		case end > 0:
			return IOError{
				Err:     err,
				Message: fmt.Sprintf("want %v bytes but only wrote %v bytes", len(buff), end),
			}
		default: // err != nil && end <= 0
			return IOError{
				Err: err,
			}
		}
	}
	return nil
}
