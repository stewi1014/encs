package encio

import "io"

type UUID [16]byte

func (buff *UUID) EncodeUUID(w io.Writer, id [2]uint64) error {
	buff[0] = uint8(id[0])
	buff[1] = uint8(id[0] >> 8)
	buff[2] = uint8(id[0] >> 16)
	buff[3] = uint8(id[0] >> 24)
	buff[4] = uint8(id[0] >> 32)
	buff[5] = uint8(id[0] >> 40)
	buff[6] = uint8(id[0] >> 48)
	buff[7] = uint8(id[0] >> 56)
	buff[8] = uint8(id[1])
	buff[9] = uint8(id[1] >> 8)
	buff[10] = uint8(id[1] >> 16)
	buff[11] = uint8(id[1] >> 24)
	buff[12] = uint8(id[1] >> 32)
	buff[13] = uint8(id[1] >> 40)
	buff[14] = uint8(id[1] >> 48)
	buff[15] = uint8(id[1] >> 56)
	return Write(buff[:], w)
}

func (buff *UUID) DecodeUUID(r io.Reader) (id [2]uint64, err error) {
	if err = Read(buff[:], r); err != nil {
		return
	}

	id[0] = uint64(buff[0])
	id[0] |= uint64(buff[1]) << 8
	id[0] |= uint64(buff[2]) << 16
	id[0] |= uint64(buff[3]) << 24
	id[0] |= uint64(buff[4]) << 32
	id[0] |= uint64(buff[5]) << 40
	id[0] |= uint64(buff[6]) << 48
	id[0] |= uint64(buff[7]) << 56
	id[1] = uint64(buff[8])
	id[1] |= uint64(buff[9]) << 8
	id[1] |= uint64(buff[10]) << 16
	id[1] |= uint64(buff[11]) << 24
	id[1] |= uint64(buff[12]) << 32
	id[1] |= uint64(buff[13]) << 40
	id[1] |= uint64(buff[14]) << 48
	id[1] |= uint64(buff[15]) << 56
	return
}
