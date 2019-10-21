package encs

import (
	"fmt"
	"io"
	"reflect"
	"unsafe"

	"github.com/stewi1014/encs/enc"
)

// NewDecoder returns a new Decoder.
func NewDecoder(r io.Reader) *Decoder {
	return &Decoder{
		r:        r,
		te:       defaultTypeEncoder,
		decoders: make(map[reflect.Type]enc.Encodable),
	}
}

// Decoder decodes data into passed values.
type Decoder struct {
	r        io.Reader
	te       enc.Resolver
	decoders map[reflect.Type]enc.Encodable
}

// Decode decodes the next value from the reader into the type referenced by v.
// v must be a pointer to the type passed to Encode.
func (d *Decoder) Decode(v interface{}) error {
	if v == nil {
		return enc.ErrNilPointer
	}

	val := reflect.ValueOf(v)
	if val.Kind() != reflect.Ptr {
		return fmt.Errorf("%v: decode must be given pointer to type", enc.ErrBadType)
	}
	if val.IsNil() {
		return enc.ErrNilPointer
	}
	val = val.Elem()
	if !val.CanSet() {
		return fmt.Errorf("%v: cannot set value of %v", enc.ErrBadType, val)
	}

	ty, err := d.te.Decode(val.Type(), d.r)
	if err != nil {
		return err
	}

	if ty != val.Type() {
		return fmt.Errorf("%v: decoding into %v, but received %v", enc.ErrBadType, val.Type(), ty)
	}

	e := d.getEncodable(ty)
	return e.Decode(unsafe.Pointer(val.UnsafeAddr()), d.r)
}

// DecodeInterface sets v to the decoded value.
func (d *Decoder) DecodeInterface(i *interface{}) error {
	if i == nil {
		return enc.ErrNilPointer
	}

	ival := reflect.ValueOf(i).Elem()
	if !ival.CanSet() {
		return fmt.Errorf("%v: cannot set value of %v", enc.ErrBadType, ival)
	}

	var elemt reflect.Type
	if !ival.IsNil() {
		elemt = ival.Elem().Type()
	}

	ty, err := d.te.Decode(elemt, d.r)
	if err != nil {
		return err
	}

	ec := d.getEncodable(ty)
	return enc.DecodeInterface(i, ec, d.r)
}

func (d *Decoder) getEncodable(t reflect.Type) enc.Encodable {
	if e, ok := d.decoders[t]; ok {
		return e
	}

	config := &enc.Config{
		Resolver: d.te,
	}

	e := enc.NewEncodable(t, config)
	d.decoders[t] = e
	return e
}
