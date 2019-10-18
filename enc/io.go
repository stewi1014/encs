package enc

import (
	"fmt"
	"io"
)

func read(b []byte, r io.Reader) error {
	end, err := r.Read(b)
	for end < len(b) {
		var n int
		n, err = r.Read(b[end:])
		end += n
		if err != nil || n == 0 {
			break
		}
	}
	if end != len(b) {
		if err == io.EOF {
			err = io.ErrUnexpectedEOF
		} else if err == nil {
			err = io.ErrNoProgress
		}

		return fmt.Errorf("%v; %v", ErrMalformed, err)
	}
	return nil
}

func write(b []byte, w io.Writer) error {
	n, err := w.Write(b)
	if n != len(b) {
		if err != nil {
			return err
		}
		return fmt.Errorf("%v; wrote %v, want %v", io.ErrShortWrite, n, len(b))
	}
	return err
}
