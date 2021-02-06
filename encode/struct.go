package encode

import (
	"errors"
	"fmt"
	"hash/crc32"
	"io"
	"reflect"
	"sort"
	"strconv"
	"unicode"
	"unsafe"

	"github.com/stewi1014/encs/encio"
)

type structMembers []reflect.StructField

func (a structMembers) Len() int           { return len(a) }
func (a structMembers) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a structMembers) Less(i, j int) bool { return a[i].Name < a[j].Name }

func structFields(ty reflect.Type) []reflect.StructField {
	fields := make(structMembers, 0, ty.NumField())
	for i := 0; i < ty.NumField(); i++ {
		field := ty.Field(i)

		tagStr, tagged := field.Tag.Lookup(StructTag)
		var tag bool
		if tagged {
			parsed, err := strconv.ParseBool(tagStr)
			if err != nil {
				fmt.Fprintf(encio.Warnings, "%v (decoding struct tag in %v)", err, ty.String())
				tagged = false
			} else {
				tag = parsed
			}
		}

		if tagged && !tag {
			// Tag says skip
			continue
		}

		if !tagged && !unicode.IsUpper([]rune(field.Name)[0]) {
			// Not tagged and not exported
			continue
		}

		fields = append(fields, field)
	}

	// Some optimisations and use cases depend on the order of fields remaining the same between systems, so we sort alphabetically.
	sort.Sort(fields)
	return fields
}

// NewStructLoose returns a new struct Encodable.
func NewStructLoose(ty reflect.Type, src Source) *StructLoose {
	if ty.Kind() != reflect.Struct {
		panic(encio.NewError(encio.ErrBadType, fmt.Sprintf("%v is not a struct", ty), 0))
	}

	e := &StructLoose{
		ty:  ty,
		len: encio.NewUint32(),
	}

	// Take a hash of each field name, generate an ID from it,
	// and populate fields with the id, Encodable and offset for the field.
	fields := structFields(ty)
	hasher := crc32.NewIEEE()
	for _, field := range fields {
		if err := encio.Write([]byte(field.Name), hasher); err != nil {
			panic(err) // hasher should never fail to write.
		}

		f := looseField{
			offset: field.Offset,
			id:     hasher.Sum32(),
			enc:    src.NewEncodable(field.Type, nil),
		}

		hasher.Reset()

		for _, existing := range e.fields {
			if existing.id == f.id {
				panic(encio.NewError(
					errors.New("hash collision"),
					fmt.Sprintf("struct field %v with type %v has same hash as previous field with type %v in struct %v", field.Name, field.Type.String(), (*existing.enc).Type(), ty.String()),
					0,
				))
			}
		}

		e.fields = append(e.fields, &f)
	}

	return e
}

// StructLoose is an Encodable for structs.
// It encodes fields by a generated ID, and maps fields with the same name.
// Fields received that do not exist locally are ignored,
// and fields that are expected that aren't received are set to their zero values.
// Exported fields can be ignored using the tag `encs:"false"`, and
// unexported fields can be included with the tag `encs:"true"`.
type StructLoose struct {
	ty     reflect.Type
	fields []*looseField
	len    encio.Uint32
}

type looseField struct {
	id     uint32
	offset uintptr
	enc    *Encodable
}

// Size implements Encodable.
func (e *StructLoose) Size() int {
	size := len(e.fields) * 4
	for _, field := range e.fields {
		fsize := (*field.enc).Size()
		if fsize < 0 {
			return -1 << 31
		}
		size += fsize
	}
	return size
}

// Type implements Encodable.
func (e *StructLoose) Type() reflect.Type { return e.ty }

// Encode implements Encodable.
func (e *StructLoose) Encode(ptr unsafe.Pointer, w io.Writer) error {
	if err := e.len.Encode(w, uint32(len(e.fields))); err != nil {
		return err
	}

	for _, field := range e.fields {
		if err := e.len.Encode(w, field.id); err != nil {
			return err
		}

		if err := (*field.enc).Encode(unsafe.Pointer(uintptr(ptr)+field.offset), w); err != nil {
			return err
		}
	}

	return nil
}

// Decode implements Encodable.
// Fields that are not received are set to their zero value.
// Fields sent that do not exist locally are ignored.
func (e *StructLoose) Decode(ptr unsafe.Pointer, r io.Reader) error {
	l, err := e.len.Decode(r)
	if err != nil {
		return err
	}
	if uintptr(l)*8 > encio.TooBig { // Don't include size of encoded struct members, way too much work here.
		return encio.NewError(encio.ErrMalformed, fmt.Sprintf("%v struct fields is too many to decode", l), 0)
	}

	var i int
	for parsed := uint32(0); parsed < l; parsed++ {
		id, err := e.len.Decode(r)
		if err != nil {
			return err
		}

		// Move next field to the current index. At the end we have all unfetched fields
		// at the end of the slice above index.
		for j := i; j < len(e.fields); j++ {
			if e.fields[j].id == id {
				e.fields[i], e.fields[j] = e.fields[j], e.fields[i]
				goto foundField
			}
		}
		// local field not found
		continue

	foundField:

		if err := (*e.fields[i].enc).Decode(unsafe.Pointer(uintptr(ptr)+e.fields[i].offset), r); err != nil {
			return err
		}

		i++
	}

	// Set undecoded fields to zero value.
	for _, field := range e.fields[i:] {
		// what a mouthful
		reflect.NewAt((*field.enc).Type(), unsafe.Pointer(uintptr(ptr)+field.offset)).Elem().Set(reflect.New((*field.enc).Type()).Elem())
	}

	return nil
}

// NewStructStrict returns a new struct Encodable.
func NewStructStrict(ty reflect.Type, src Source) *StructStrict {
	if ty.Kind() != reflect.Struct {
		panic(encio.NewError(encio.ErrBadType, fmt.Sprintf("%v is not a struct", ty), 0))
	}

	fields := structFields(ty)
	s := &StructStrict{
		ty:     ty,
		fields: make([]strictField, len(fields)),
	}

	for i := range fields {
		s.fields[i].offset = fields[i].Offset
		s.fields[i].enc = src.NewEncodable(fields[i].Type, nil)
	}

	return s
}

// StructStrict is an Encodable for structs.
// It encodes fields in a determenistic order with no ID.
// There must be no discrepancy in the struct's type between Encode and Decode.
// Use Type to check types if unsure of exact simmilarity.
// Exported fields can be ignored using the tag `encs:"false"`, and
// unexported fields can be included with the tag `encs:"true"`.
type StructStrict struct {
	ty     reflect.Type
	fields []strictField
}

type strictField struct {
	offset uintptr
	enc    *Encodable
}

// Size implements Encodable.
func (e StructStrict) Size() (size int) {
	for _, member := range e.fields {
		msize := (*member.enc).Size()
		if msize < 0 {
			return -1 << 31
		}
		size += msize
	}
	return
}

// Type implements Encodable.
func (e StructStrict) Type() reflect.Type { return e.ty }

// Encode implements Encodable.
func (e StructStrict) Encode(ptr unsafe.Pointer, w io.Writer) error {
	checkPtr(ptr)
	for _, m := range e.fields {
		if err := (*m.enc).Encode(unsafe.Pointer(uintptr(ptr)+m.offset), w); err != nil {
			return err
		}
	}
	return nil
}

// Decode implements Encodable.
func (e StructStrict) Decode(ptr unsafe.Pointer, r io.Reader) error {
	// I really don't like doubling up on the Struct encodable, but wow,
	// what a difference it makes.
	checkPtr(ptr)
	for _, m := range e.fields {
		if err := (*m.enc).Decode(unsafe.Pointer(uintptr(ptr)+m.offset), r); err != nil {
			return err
		}
	}
	return nil
}
