package encio

import "io"

const (
	maxSingleUint = 255 - 4
)

// Uvarint provides fast methods for reading and writing uint32s in variable-length format.
type Uvarint [5]byte

// Encode writes n to w
func (buff *Uvarint) Encode(w io.Writer, n uint32) error {
	if n < maxSingleUint {
		buff[0] = uint8(n)
		return Write(buff[:1], w)
	}
	size := uint8(1)
	for n > 0 {
		buff[size] = uint8(n)
		n >>= 8
		size++
	}
	buff[0] = maxSingleUint + size - 1
	return Write(buff[:size], w)
}

// Decode reads a uint32 from r
func (buff *Uvarint) Decode(r io.Reader) (uint32, error) {
	if err := Read(buff[:1], r); err != nil {
		return 0, err
	}
	if buff[0] < maxSingleUint {
		return uint32(buff[0]), nil
	}
	size := buff[0] - maxSingleUint
	if err := Read(buff[:size], r); err != nil {
		return 0, err
	}
	n := uint32(0)
	for i := byte(0); i < size; i++ {
		n |= uint32(buff[i]) << (i * 8)
	}
	return n, nil
}
