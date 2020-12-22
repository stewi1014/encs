package encio

import (
	"crypto/rand"
	"io"
	"math/bits"
)

// NewUUID returns a randomly generated UUIDv4.
func NewUUID() UUID {
	var uuid UUID
	n, err := rand.Read(uuid[:])
	if n != 16 {
		if err != nil {
			panic(err)
		}
		panic("Short UUID read")
	}

	uuid[6] = (uuid[6] & 0x0f) | 0x40 // major version
	uuid[8] = (uuid[8] & 0x3f) | 0x80 // minor version

	return uuid
}

// UUID is a Universally Unique Identifier.
type UUID [16]byte

// NewUint32 returns a Uint32.
func NewUint32() Uint32 {
	return Uint32{
		buff: make([]byte, 4),
	}
}

// Uint32 provides methods for encoding and decoding uint32s.
type Uint32 struct {
	buff []byte
}

// Encode writes the given uint32 to w.
func (e *Uint32) Encode(w io.Writer, n uint32) error {
	EncodeUint32(e.buff, n)
	return Write(e.buff, w)
}

// Decode decodes a uint32 from r.
func (e *Uint32) Decode(r io.Reader) (uint32, error) {
	err := Read(e.buff, r)
	return DecodeUint32(e.buff), err
}

// EncodeUint32 writes a uint32 to buff.
func EncodeUint32(buff []byte, n uint32) {
	buff[0] = uint8(n)
	buff[1] = uint8(n >> 8)
	buff[2] = uint8(n >> 16)
	buff[3] = uint8(n >> 24)
}

// DecodeUint32 reads a uint32 from buff.
func DecodeUint32(buff []byte) uint32 {
	n := uint32(buff[0])
	n |= uint32(buff[1]) << 8
	n |= uint32(buff[2]) << 16
	n |= uint32(buff[3]) << 24
	return n
}

// NewInt32 returns a new Int32.
func NewInt32() Int32 {
	return Int32{
		buff: make([]byte, 4),
	}
}

// Int32 provides methods for encoding int32s.
type Int32 struct {
	buff []byte
}

// Encode writes the given int32 to w.
func (e *Int32) Encode(w io.Writer, n int32) error {
	EncodeInt32(e.buff, n)
	return Write(e.buff, w)
}

// Decode decodes a int32 from r.
func (e *Int32) Decode(r io.Reader) (int32, error) {
	err := Read(e.buff, r)
	return DecodeInt32(e.buff), err
}

// EncodeInt32 writes an int32 to buff.
func EncodeInt32(buff []byte, n int32) {
	buff[0] = uint8(n)
	buff[1] = uint8(n >> 8)
	buff[2] = uint8(n >> 16)
	buff[3] = uint8(n >> 24)
}

// DecodeInt32 reads a int32 from buff.
func DecodeInt32(buff []byte) int32 {
	n := int32(buff[0])
	n |= int32(buff[1]) << 8
	n |= int32(buff[2]) << 16
	n |= int32(buff[3]) << 24
	return n
}

// NewVaruint32 returns a Varuint for encoding variable-length encoded uint32s.
func NewVaruint32() Varuint32 {
	return Varuint32{
		buff: make([]byte, 4),
	}
}

// Varuint32 provides methods for encoding uint32s with variable-length encoding.
// It can only encode uint32s up to 1<<30 - 1.
type Varuint32 struct {
	buff []byte
}

// Encode writes the given uint32 to w.
// It returns the number of bytes written, and any write errors.
func (e *Varuint32) Encode(w io.Writer, n uint32) (int, error) {
	l := EncodeVarUint32(e.buff, n)
	return l, Write(e.buff[:l], w)
}

// Decode decodes a uint32 from r.
func (e *Varuint32) Decode(r io.Reader) (uint32, error) {
	err := Read(e.buff[:1], r)
	if err != nil {
		return 0, err
	}
	n, size := DecodeVarUint32Header(e.buff[0])
	if size != 0 {
		err = Read(e.buff[1:size+1], r)
		n, _ = DecodeVarUint32(e.buff)
	}

	return n, err
}

const (
	// MaxVarint is the largest uint32 that can be encoded with EncodeVarUint32.
	MaxVarint = uint32(1<<30 - 1)
	sizeMask  = (1<<2 - 1)
)

// EncodeVarUint32 encodes the given uint32 in variable-length format to buff. It returns the encoded length.
// It's primary use case is writing the length of a following message.
//
// EncodeVarUint32 is fast to encode, especially for small numbers, and only writes up to 4 bytes, but at the cost of the 2 most sigificant bits in n.
// They are not encoded; n must not be larger than MaxVarint.
//
// Buff must be large enough to write the int, as a general rule, it should be 4 bytes large.
func EncodeVarUint32(buff []byte, n uint32) int {
	n <<= 2
	n |= uint32(bits.Len32(n>>1) / 8)
	switch n & sizeMask {
	case 0:
		buff[0] = byte(n)
		return 1
	case 1:
		buff[0] = byte(n)
		buff[1] = byte(n >> 8)
		return 2
	case 2:
		buff[0] = byte(n)
		buff[1] = byte(n >> 8)
		buff[2] = byte(n >> 16)
		return 3
	case 3:
		buff[0] = byte(n)
		buff[1] = byte(n >> 8)
		buff[2] = byte(n >> 16)
		buff[3] = byte(n >> 24)
		return 4
	default:
		panic("impossible")
	}
}

// DecodeVarUint32Header decodes the header sent by EncodeVarUint32,
// returning either how many bytes are needed to decode, or the encoded uint32.
// If size is 0, n is the decoded number, if not, subsequent calls to DecodeVarUint32 must include the header again.
func DecodeVarUint32Header(b byte) (n uint32, size int) {
	if b&sizeMask == 0 {
		return uint32(b >> 2), 0
	}
	return 0, int(b & sizeMask)
}

// DecodeVarUint32 decodes a uint32 from the given buffer.
// It decodes the header the same way as DecodeVarUint32Header, however it goes ahead and reads more data if it's needed,
// always returning the decoded number.
func DecodeVarUint32(buff []byte) (n uint32, size int) {
	switch buff[0] & sizeMask {
	case 0:
		return uint32(buff[0] >> 2), 1
	case 1:
		n = uint32(buff[0]) >> 2
		n |= uint32(buff[1]) << 6
		return n, 2
	case 2:
		n = uint32(buff[0]) >> 2
		n |= uint32(buff[1]) << 6
		n |= uint32(buff[2]) << 14
		return n, 3
	case 3:
		n = uint32(buff[0]) >> 2
		n |= uint32(buff[1]) << 6
		n |= uint32(buff[2]) << 14
		n |= uint32(buff[3]) << 22
		return n, 4
	default:
		panic("impossible")
	}
}
