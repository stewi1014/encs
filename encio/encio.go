// Package encio provides io methods relevant to encoding, as well as error types.
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
	TooBig = uintptr(1 << (25 + ((^uint(0) >> 32) & 2)))
)

// Read reads from r, completely filling the buffer. It provides error handling with as little overhead as possible.
// In an ideal read, only a single int equality check is performed. If the read reports the whole buffer is read, returned errors are ignored.
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
			return NewIOError(
				errors.New("bad io.Reader implementation"),
				r,
				fmt.Sprintf("reported %v bytes read, but buffer is only %v bytes", end, len(buff)),
				1,
			)
		case errors.Is(err, io.EOF):
			return NewIOError(
				io.ErrUnexpectedEOF,
				r,
				fmt.Sprintf("want %v bytes but only got %v", len(buff), end),
				1,
			)
		case err != nil:
			return err
		default: // err == nil
			return NewIOError(
				io.ErrNoProgress,
				r,
				fmt.Sprintf("want %v bytes but only got %v", end, len(buff)),
				1,
			)
		}
	}
	return nil
}

// Write writes to w from buff, handling errors of io.Writer with as little overhead as possible.
// In an ideal write, only a single int equality check is performed. It returns any error from Write().
func Write(buff []byte, w io.Writer) error {
	n, err := w.Write(buff)
	if n == len(buff) {
		return err
	}

	end := n
	for end < len(buff) && err == nil && n > 0 {
		fmt.Fprintf(Warnings, "encs: %T is a bad io.Writer implementation. It wrote short (given %v bytes but reported only %v written) yet returned no error. Will call it again...\n", w, len(buff)-(end-n), n)
		n, err = w.Write(buff[end:])
		end += n
	}

	if end != len(buff) {
		switch {
		case end > len(buff):
			return NewIOError(
				errors.New("bad io.Writer implementation"),
				w,
				fmt.Sprintf("Write() reported %v bytes written, but was only given %v bytes", end, len(buff)),
				1,
			)
		case err == nil:
			return NewIOError(
				io.ErrShortWrite,
				w,
				fmt.Sprintf("want %v bytes but only wrote %v bytes", len(buff), end),
				1,
			)
		default:
			return NewIOError(
				err,
				w,
				fmt.Sprintf("want %v bytes but wrote %v bytes", len(buff), end),
				1,
			)
		}
	}
	return nil
}
