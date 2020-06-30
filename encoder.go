package encs

import (
	"io"
	"reflect"
	"unsafe"

	"github.com/stewi1014/encs/encio"
	"github.com/stewi1014/encs/encodable"
)

func NewEncoder(w io.Writer, config *Config) *Encoder {
	config = config.copyAndFill()
	return &Encoder{
		w:        w,
		resolver: config.Resolver,
		source: encodable.NewSource(&encodable.Config{
			Resolver: config.Resolver,
		}, encodable.New),
	}
}

type Encoder struct {
	w        io.Writer
	resolver encodable.Resolver
	source   *encodable.Source
}

func (e *Encoder) Encode(v interface{}) error {
	if v == nil {
		return encio.NewError(encio.ErrNilPointer, "cannot encode nil interface", 0)
	}

	t := reflect.TypeOf(v)
	if t.Kind() != reflect.Ptr {
		return encio.NewError(encio.ErrBadType, "values must be passed by reference", 0)
	}

	t = t.Elem() // the encoded type

	err := e.resolver.Encode(t, e.w)
	if err != nil {
		return err
	}

	ec := e.source.GetEncodable(t)
	// we've already confirmed that the interface contains a pointer type,
	// elem should be a pointer to the actual value, not a pointer; the pointer type seems to be stored internally in the interface,
	// so we take just take the address and go.
	return ec.Encode(ptrInterface(unsafe.Pointer(&v)).elem, e.w)
}

/*

func NewEncoder(w io.Writer) *Encoder {
	return &Encoder{
		w:        w,
		te:       defaultTypeEncoder,
		encoders: make(map[reflect.Type]enc.Encodable),
	}
}

type Encoder struct {
	w        io.Writer
	te       enc.Resolver
	encoders map[reflect.Type]enc.Encodable
}

func (e *Encoder) Encode(v interface{}) error {
	if v == nil {
		return enc.ErrNilPointer
	}
	val := reflect.ValueOf(v)

	err := e.te.Encode(val.Type(), e.w)
	if err != nil {
		return err
	}

	ec := e.getEncodable(val.Type())
	return enc.Encode(ec, v, e.w)
}

func (e *Encoder) getEncodable(t reflect.Type) enc.Encodable {
	if ec, ok := e.encoders[t]; ok {
		return ec
	}

	config := &enc.Config{
		Resolver: e.te,
	}

	ec := enc.NewEncodable(t, config)
	e.encoders[t] = ec
	return ec
}
*/
