package gneg

import (
	"context"
	"io"
	"reflect"

	"github.com/stewi1014/gneg/gram"
)

// NewDecoder returns a new Decoder on r.
func NewDecoder(r io.Reader, config *Config) *Decoder {
	d := &Decoder{
		typeDecoders: make(map[reflect.Type]etype),
	}

	if config == nil || config.GramDecoder == nil {
		d.gramReader = gram.NewStreamReader(r)
	} else {
		d.gramReader = config.GramDecoder
	}

	if config == nil || config.TypeResolver == nil {
		d.resolver = NewCachingResolver(defaultResolver)
	} else {
		d.resolver = config.TypeResolver
	}

	return d
}

// Decoder decodes types
type Decoder struct {
	gramReader   gram.Reader
	resolver     TypeResolver
	typeDecoders map[reflect.Type]etype
}

// ContextDecode decodes with a context.
func (d *Decoder) ContextDecode(ctx context.Context, v interface{}) error {
	var derr error
	done := make(chan struct{})
	go func() {
		derr = d.Decode(v)
		close(done)
	}()
	select {
	case <-done:
		return derr
	case <-ctx.Done():
		return ctx.Err()
	}
}

// Decode decodes into val
func (d *Decoder) Decode(v interface{}) error {
	g, err := d.gramReader.Read()
	if err != nil {
		return err
	}

	l := g.ReadUint32()
	ty, err := d.resolver.Decode(g.LimitReader(int(l)))
	if err != nil {
		return err
	}

	et, ok := d.typeDecoders[ty]
	if !ok {
		et, err = newetype(ty)
		if err != nil {
			return err
		}
		d.typeDecoders[ty] = et
	}

	val := reflect.ValueOf(v).Elem()
	if val.Kind() == reflect.Interface {
		return decodeIntoInterface(val, et, g)
	}
	return et.Decode(val, g)
}

// correctly initialise and decode into an interface.
func decodeIntoInterface(v reflect.Value, et etype, g *gram.Gram) error {
	if v.IsNil() || et.EncType() != v.Elem().Type() {
		n := reflect.New(et.EncType()).Elem()
		err := et.Decode(n, g)
		if err != nil {
			return err
		}

		v.Set(n)
		return nil
	}
	return et.Decode(v.Elem(), g)
}
