package gneg

import (
	"io"
	"reflect"

	"github.com/stewi1014/gneg/gram"
)

// NewEncoder returns a new encoder writing to w.
// config can be nil.
func NewEncoder(w io.Writer, config *Config) *Encoder {
	e := &Encoder{
		typeEncoders: make(map[reflect.Type]etype),
	}

	// Gram Writer
	if config == nil || config.GramEncoder == nil {
		e.gramWriter = gram.NewStreamWriter(w)
	} else {
		e.gramWriter = config.GramEncoder
	}

	// Type Encoder
	if config == nil || config.TypeResolver == nil {
		e.resolver = NewCachingResolver(defaultResolver)
	} else {
		e.resolver = config.TypeResolver
	}

	return e
}

// Encoder provides a method for encoding data into a stream.
type Encoder struct {
	gramWriter   gram.Writer
	resolver     TypeResolver
	typeEncoders map[reflect.Type]etype
}

// Encode encodes v.
func (e *Encoder) Encode(v interface{}) error {
	val := reflect.ValueOf(v)
	ty := val.Type()
	g, write := e.gramWriter.Write()

	header := gram.WriteSizeHeader(g)
	err := e.resolver.Encode(ty, g)
	if err != nil {
		return err
	}
	header()

	et, ok := e.typeEncoders[ty]
	if !ok {
		et, err = newetype(ty)
		if err != nil {
			return err
		}
		e.typeEncoders[ty] = et
	}

	err = et.Encode(val, g)
	if err != nil {
		return err
	}

	return write()
}
