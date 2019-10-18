package encs

import (
	"fmt"
	"io"
	"reflect"
	"unsafe"

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
	te       enc.Type
	encoders map[reflect.Type]enc.Encodable
}

func (e *Encoder) Encode(v interface{}) error {
	if v == nil {
		return enc.ErrNilPointer
	}
	val := reflect.ValueOf(v)
	if val.Kind() != reflect.Ptr {
		return fmt.Errorf("%v: values must be passed by reference", enc.ErrBadType)
	}
	if val.IsNil() {
		return enc.ErrNilPointer
	}
	val = val.Elem()
	if !val.CanAddr() {
		return fmt.Errorf("%v: cannot get address of %v", enc.ErrBadType, val)
	}

	err := e.te.Encode(val.Type(), e.w)
	if err != nil {
		return err
	}

	ec := e.getEncodable(val.Type())
	return ec.Encode(unsafe.Pointer(val.UnsafeAddr()), e.w)
}

func (e *Encoder) getEncodable(t reflect.Type) enc.Encodable {
	if ec, ok := e.encoders[t]; ok {
		return ec
	}

	config := &enc.Config{
		TypeEncoder: e.te,
	}

	ec := enc.NewEncodable(t, config)
	e.encoders[t] = ec
	return ec
}
