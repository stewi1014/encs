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
// In an ideaĺ read, only a single int equality check is performed. If the read reports the whole buffer is read, returned errors are ignored.
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
				errors.New("oversized read"),
				r,
				fmt.Sprintf("reported %v bytes read, but buffer is only %v bytes", end, len(buff)),
				0,
			)
		case err == io.EOF:
			return NewIOError(
				err,
				r,
				fmt.Sprintf("want %v bytes but only got %v", len(buff), end),
				0,
			)
		case err != nil:
			return NewIOError(
				err,
				r,
				"",
				0,
			)
		default: //err == nil
			return NewIOError(
				io.ErrNoProgress,
				r,
				fmt.Sprintf("read %v bytes, need %v bytes", end, len(buff)),
				0,
			)
		}
	}
	return nil
}

// Write writes to w from buff, handling errors of io.Writer with little overhead in almost all cases.
// In an ideal write, only a single int equality check is performed. It returns any error from Write().
func Write(buff []byte, w io.Writer) error {
	n, err := w.Write(buff)
	if n == len(buff) {
		return err
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
			return NewIOError(
				errors.New("bad io.Writer implementation"),
				w,
				fmt.Sprintf("Write() reported %v bytes written, but was only given %v bytes", end, len(buff)),
				0,
			)
		case err == nil:
			return NewIOError(
				io.ErrShortWrite,
				w,
				fmt.Sprintf("want %v bytes but only wrote %v bytes", len(buff), end),
				0,
			)
		default:
			return NewIOError(
				err,
				w,
				fmt.Sprintf("want %v bytes but wrote %v bytes", len(buff), end),
				0,
			)
		}
	}
	return nil
}
