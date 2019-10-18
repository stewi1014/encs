package enc

import (
	"fmt"
	"io"
	"reflect"
	"unsafe"
)

// callers *must* return Encodable as their new Encodable.
func (c *Config) referencer(enc Encodable) (*referencer, Encodable) {
	if c.r == nil {
		c.r = &referencer{
			encoders: make(map[reflect.Type]*ConcurrentEncodable),
			enc:      enc,
		}
		return c.r, c.r
	}
	return c.r, enc
}

// referencer resolves recursive types and values, and stops re-encoding of pointers to the same value.
type referencer struct {
	// Encodable we're wrapping
	// it doesn't need to be top-level, just a member Encodable that will receive calls to Encode() and Decode() before
	// any calls will be made to encode/decodeReference. Simply wrapping the first Encodable to ask to have a referencer should suffice.
	enc      Encodable
	encoders map[reflect.Type]*ConcurrentEncodable
	buff     [1]byte

	referenceByIndex map[uint]unsafe.Pointer
	indexByReference map[unsafe.Pointer]uint
	index            uint

	uintEnc Uint
}

const (
	refNil = 1 << iota
	refReference
	refEncoded
)

// newEncodable returns the concurrent-safe encodable for the type t, creating if needed.
// that is, a recursive-safe Encodable. Recursive types *must* use this instead of NewEncodable when resolving
// their element types, else they loop to infinity.
func (ref *referencer) newEncodable(t reflect.Type, c *Config) Encodable {
	if enc, ok := ref.encoders[t]; ok {
		return enc
	}

	enc := newConcurrentEncodable(t, c)
	ref.encoders[t] = enc
	return enc
}

// encodeReference encodes the object at ptr once, writing references to the first encode in subsequent calls to encodeReference.
// Writes up to 10 bytes.
// It is acceptable to pass a nil elem encodable with a nil pointer
func (ref *referencer) encodeReference(ptr unsafe.Pointer, elem Encodable, w io.Writer) error {
	if ptr == nil {
		ref.buff[0] = refNil
		return write(ref.buff[:], w)
	}

	if index, seen := ref.indexByReference[ptr]; seen {
		ref.buff[0] = refReference
		if err := write(ref.buff[:], w); err != nil {
			return err
		}
		return ref.uintEnc.Encode(unsafe.Pointer(&index), w)
	}

	ref.index++
	ref.indexByReference[ptr] = ref.index
	ref.buff[0] = refEncoded
	if err := write(ref.buff[:], w); err != nil {
		return err
	}

	return elem.Encode(ptr, w)
}

// decodeReference does the opposite of encodeReference, pointing ptr to the decoded object.
func (ref *referencer) decodeReference(ptr *unsafe.Pointer, elem Encodable, r io.Reader) error {
	if ptr == nil {
		return ErrNilPointer
	}

	if err := read(ref.buff[:], r); err != nil {
		return err
	}

	switch ref.buff[0] {
	case refNil:
		*ptr = nil
		return nil
	case refReference:
		var index uint
		err := ref.uintEnc.Decode(unsafe.Pointer(&index), r)
		if err != nil {
			return err
		}
		var ok bool
		*ptr, ok = ref.referenceByIndex[index]
		if !ok {
			return fmt.Errorf("%v: unknown reference index, bad metadata", ErrMalformed)
		}
		return nil
	case refEncoded:
		break

	default:
		return fmt.Errorf("%v: unknown reference descriptor, bad metadata", ErrMalformed)
	}

	// we must decode the type, and store a pointer to it
	if *ptr == nil {
		newAt(ptr, elem.Type())
	}

	ref.index++
	ref.referenceByIndex[ref.index] = *ptr

	return elem.Decode(*ptr, r)

}

// referencer must know when each encode and decode ends. less we make all meta-type Encodables reset us every encode/decode,
// we make referencer wrap the Encodable.

func (ref *referencer) Encode(ptr unsafe.Pointer, w io.Writer) error {
	ref.indexByReference = make(map[unsafe.Pointer]uint)
	ref.index = 0
	return ref.enc.Encode(ptr, w)
}

func (ref *referencer) Decode(ptr unsafe.Pointer, r io.Reader) error {
	ref.referenceByIndex = make(map[uint]unsafe.Pointer)
	ref.index = 0
	return ref.enc.Decode(ptr, r)
}

func (ref *referencer) Type() reflect.Type {
	return ref.enc.Type()
}

func (ref *referencer) Size() int {
	return ref.enc.Size()
}

/*
func newRecursiveParent(elem Encodable) *pointerParent {
	rp := &pointerParent{
		elem: elem,
		buff: make([]byte, 5),
	}
	if i, ok := elem.(initialiser); ok {
		i.init(rp)
	}
	return rp
}

// pointerParent wraps a type with pointers, encoding the pointers
type pointerParent struct {
	// both
	indexes []unsafe.Pointer
	buff    []byte
	off     int
	elem    Encodable
}

func (p *pointerParent) Size() int {
	if sized, ok := p.elem.(Sized); ok {
		return sized.Size()
	}
	return -1 << 31
}

func (p *pointerParent) Type() reflect.Type {
	return p.elem.Type()
}

func (p *pointerParent) Encode(ptr unsafe.Pointer, w io.Writer) error {
	p.w = w
	p.off = 0
	p.indexes = p.indexes[:0]
	return p.elem.Encode(ptr, p)
}

func (p *pointerParent) Decode(ptr unsafe.Pointer, r io.Reader) error {
	p.r = r
	p.off = 0
	p.indexes = p.indexes[:0]
	return p.elem.Decode(ptr, p)
}

// newEncoded notifies the parent that the object at the address ptr is being encoded at the current buffer index.
// if the object has already been encoded, seen is true.
// It writes up to 5 bytes (1 or 5).
func (p *pointerParent) newEncoded(ptr unsafe.Pointer) (seen bool, err error) {
	for i, ep := range p.indexes {
		if ep == ptr {
			seen = true
			err = p.writeLink(uint32(i))
			return
		}
	}

	p.buff[0] = firstInstance
	err = write(p.buff[:1], p)
	return
}

func (p *pointerParent) writeLink(index uint32) error {
	p.buff[0] = linked
	p.buff[1] = uint8(index)
	p.buff[2] = uint8(index >> 8)
	p.buff[3] = uint8(index >> 16)
	p.buff[4] = uint8(index >> 24)

	return write(p.buff, p)
}

// newDecoded attempts to read a link from the buffer. If successful, the decoder can point to the returned address instead of decoding.
// If it's the first instance of the value, firstDecode will be non-nil, the value should be decoded, and the address passed to finishDecode.
func (p *pointerParent) newDecoded() (ptr unsafe.Pointer, firstDecode func(unsafe.Pointer), err error) {
	err = read(p.buff[:1], p)
	if err != nil {
		return
	}

	if p.buff[0] == firstInstance {
		firstDecode = func(ptr unsafe.Pointer) {
			p.indexes = append(p.indexes, ptr)
		}
		return
	}

	err = read(p.buff[:4], p)
	if err != nil {
		return
	}

	i := uint32(p.buff[0])
	i |= uint32(p.buff[1]) << 8
	i |= uint32(p.buff[2]) << 16
	i |= uint32(p.buff[3]) << 24
	if i >= uint32(len(p.indexes)) {
		err = fmt.Errorf("%v: given link index is out of bounds", ErrMalformed)
		return
	}

	ptr = p.indexes[i]
	return
}
*/
