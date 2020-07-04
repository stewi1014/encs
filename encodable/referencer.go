package encodable

import (
	"io"
	"reflect"
	"unsafe"

	"github.com/stewi1014/encs/encio"
)

// TODO: Throw all of this in the trash. It's slow, complex and can be done smarter

// referencer resolves recursive types and values, and stops re-encoding of pointers to the same value.
// member Encodables call newEncodable, encodeReference and decodeReference instead of performing their own sub-type encoding.
type referencer struct {
	// Encodable we're wrapping
	// it doesn't need to be top-level, just a member Encodable that will receive calls to Encode() and Decode() before
	// any calls will be made to encode/decodeReference. Simply wrapping the first Encodable to ask to have a referencer should suffice.
	enc Encodable

	// index is a unique id for an encoded reference type. indexes are not static, and are resolved on every decode and encode.
	// for encoding, this is used to ensure the type at a given pointer is only encoded once, with subsequent encodes only writing a link (index) to the previously encoded value to the buffer.
	// for decoding, this is used to keep track of decoded types, and to resolve links (index) to previously decoded values when they are read from the buffer.
	references []unsafe.Pointer

	// encoders is a map of Encodables for given types. used to prevent infinite recursion when resolving recursive types.
	encoders map[reflect.Type]*Concurrent
	buff     [1]byte

	// intEnc is used for encoding the index.
	intEnc Int
}

const (
	refNil = 1 << iota
	refReference
	refEncoded
)

// newEncodable returns the concurrent-safe encodable for the type t, creating if needed.
// that is, a recursive-safe Encodable. Recursive types *must* use this instead of NewEncodable when resolving
// their element types, else they loop to infinity.
// also, they must be concurrent safe on the off chance that a type embeds itself, and calls itself from inside Encode and Decode.
func (ref *referencer) newEncodable(t reflect.Type, state *state) Encodable {
	if enc, ok := ref.encoders[t]; ok {
		return enc
	}

	enc := NewConcurrent(func() Encodable {
		return newEncodable(t, state)
	})
	ref.encoders[t] = enc
	return enc
}

// encodeReference encodes the object at ptr once, writing references to the first encode in subsequent calls to encodeReference.
// Writes up to 10 bytes.
// It is acceptable to pass a nil elem encodable with a nil pointer (but Encodables should know their sub-types beforehand).
func (ref *referencer) encodeReference(ptr unsafe.Pointer, elem Encodable, w io.Writer) error {
	if ptr == nil {
		ref.buff[0] = refNil
		return encio.Write(ref.buff[:], w)
	}

	if index, seen := ref.findPtr(ptr); seen {
		ref.buff[0] = refReference
		if err := encio.Write(ref.buff[:], w); err != nil {
			return err
		}
		return ref.intEnc.Encode(unsafe.Pointer(&index), w)
	}

	ref.append(ptr)

	ref.buff[0] = refEncoded
	if err := encio.Write(ref.buff[:], w); err != nil {
		return err
	}

	return elem.Encode(ptr, w)
}

// decodeReference does the opposite of encodeReference, pointing ptr to the decoded object.
func (ref *referencer) decodeReference(ptr *unsafe.Pointer, elem Encodable, r io.Reader) error {
	if ptr == nil {
		return encio.ErrNilPointer
	}

	if err := encio.Read(ref.buff[:], r); err != nil {
		return err
	}

	switch ref.buff[0] {
	case refNil:
		*ptr = nil
		return nil
	case refReference:
		var index int
		err := ref.intEnc.Decode(unsafe.Pointer(&index), r)
		if err != nil {
			return err
		}
		if index >= len(ref.references) {
			return encio.IOError{
				Err:     encio.ErrMalformed,
				Message: "object is stored by reference, but the referenced location doesnt exist",
			}
		}

		*ptr = ref.references[index]
		return nil
	case refEncoded:
		break

	default:
		return encio.IOError{
			Err:     encio.ErrMalformed,
			Message: "reference type byte is not nil, reference or encoded",
		}
	}

	// we must decode the type, and store a pointer to it
	if *ptr == nil {
		value := reflect.New(elem.Type())
		*ptr = unsafe.Pointer(value.Pointer())
	}

	ref.append(*ptr) // Must be before elem.Decode in case it calls us during its decode.
	return elem.Decode(*ptr, r)
}

func (ref *referencer) findPtr(ptr unsafe.Pointer) (index int, ok bool) {
	for i, p := range ref.references {
		if p == ptr {
			return i, true
		}
	}
	return 0, false
}

func (ref *referencer) append(ptr unsafe.Pointer) {
	l := len(ref.references)
	c := cap(ref.references)
	if l < c {
		ref.references = ref.references[:l+1]
		ref.references[l] = ptr
		return
	}
	if c == 0 {
		c = 8
	}
	nb := make([]unsafe.Pointer, c*2)
	copy(ref.references, nb)
	ref.references = nb[:l+1]
	ref.references[l] = ptr
}

// referencer must know when each encode and decode ends. less we make all reference or compund type Encodables reset us every encode/decode,
// in which case the top-level Encodable would have to know it's the parent, and act differently;
// we make referencer wrap the first Encodable to request our presence.
// This relies on the fact that compund Encodable types will always execute child encodables in the same order.

func (ref *referencer) String() string {
	return "(referencer)" + ref.enc.String()
}

func (ref *referencer) Size() int {
	return ref.enc.Size()
}

func (ref *referencer) Type() reflect.Type {
	return ref.enc.Type()
}

func (ref *referencer) Encode(ptr unsafe.Pointer, w io.Writer) error {
	ref.references = ref.references[:0]
	return ref.enc.Encode(ptr, w)
}

func (ref *referencer) Decode(ptr unsafe.Pointer, r io.Reader) error {
	ref.references = ref.references[:0]
	return ref.enc.Decode(ptr, r)
}
