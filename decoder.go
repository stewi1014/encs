package encs

import (
	"fmt"
	"io"
	"reflect"
	"unsafe"

	"github.com/stewi1014/encs/encio"
	"github.com/stewi1014/encs/encodable"
)

func NewDecoder(r io.Reader, config *Config) *Decoder {
	config = config.copyAndFill()
	return &Decoder{
		r:        r,
		resolver: config.Resolver,
		source: encodable.NewSource(&encodable.Config{
			Resolver: config.Resolver,
		}, encodable.New),
	}
}

type Decoder struct {
	r        io.Reader
	resolver encodable.Resolver
	source   *encodable.Source
}

func (d *Decoder) Decode(v interface{}) error {
	if v == nil {
		return encio.NewError(encio.ErrNilPointer, "cannot decode into nil interface", 0)
	}

	val := reflect.ValueOf(v)
	if val.Kind() != reflect.Ptr {
		return encio.NewError(encio.ErrBadType, fmt.Sprintf("decoded values must be passed by reference (pointer), got %v", val.Type()), 0)
	}
	if val.IsNil() {
		return encio.NewError(encio.ErrNilPointer, "cannot decode into nil pointer", 0)
	}

	val = val.Elem()
	if !val.CanSet() {
		return encio.NewError(encio.ErrBadType, fmt.Sprintf("%v is not mutable", val.Type()), 0)
	}

	ty, err := d.resolver.Decode(val.Type(), d.r)
	if err != nil {
		return err
	}

	if ty != val.Type() {
		return encio.NewError(encio.ErrBadType, fmt.Sprintf("cannot set %v to received type %v", val.Type(), ty), 0)
	}

	return d.source.GetEncodable(ty).Decode(unsafe.Pointer(val.UnsafeAddr()), d.r)
}
