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
// also, they must be concurrent safe on the off chance that a type embeds itself, in which case it will call itself from inside Encode and Decode.
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
