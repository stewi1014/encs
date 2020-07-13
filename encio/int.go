package encio

import "io"

// Uint provides fast methods for reading and writing uint32s.
type Uint [4]byte

// EncodeUint32 writes n to w.
func (buff *Uint) EncodeUint32(w io.Writer, n uint32) error {
	buff[0] = uint8(n)
	buff[1] = uint8(n >> 8)
	buff[2] = uint8(n >> 16)
	buff[3] = uint8(n >> 24)
	return Write(buff[:], w)
}

// DecodeUint32 reads a uint32 from r.
func (buff *Uint) DecodeUint32(r io.Reader) (uint32, error) {
	if err := Read(buff[:], r); err != nil {
		return 0, err
	}
	n := uint32(buff[0])
	n |= uint32(buff[1]) << 8
	n |= uint32(buff[2]) << 16
	n |= uint32(buff[3]) << 24
	return n, nil
}

// Int provides fast methods for reading and writing int32s.
type Int [4]byte

// EncodeInt32 writes n to w.
func (buff *Int) EncodeInt32(w io.Writer, n int32) error {
	buff[0] = uint8(n)
	buff[1] = uint8(n >> 8)
	buff[2] = uint8(n >> 16)
	buff[3] = uint8(n >> 24)
	return Write(buff[:], w)
}

// DecodeInt32 reads a int32 from r.
func (buff *Int) DecodeInt32(r io.Reader) (int32, error) {
	if err := Read(buff[:], r); err != nil {
		return 0, err
	}
	n := int32(buff[0])
	n |= int32(buff[1]) << 8
	n |= int32(buff[2]) << 16
	n |= int32(buff[3]) << 24
	return n, nil
}
