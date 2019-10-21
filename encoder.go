package encs

import (
	"io"
	"reflect"

	"github.com/stewi1014/encs/enc"
)

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
	return enc.EncodeInterface(v, ec, e.w)
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
