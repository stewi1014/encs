package encs

import (
	"fmt"
	"io"
	"reflect"
	"sync"
	"unsafe"

	"github.com/stewi1014/encs/encio"
	"github.com/stewi1014/encs/encodable"
	"github.com/stewi1014/encs/encode"
)

func NewDecoder(r io.Reader) *Decoder {
	return &Decoder{
		r:       r,
		typeEnc: encode.NewType(false),
		source:  encodable.NewCachingSource(encode.NewRecursiveSource(DefaultSource)),
	}
}

type Decoder struct {
	r       io.Reader
	mutex   sync.Mutex
	typeEnc *encode.Type
	source  encodable.Source
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

	d.mutex.Lock()
	defer d.mutex.Unlock()

	ty := val.Type()
	err := d.typeEnc.Decode(unsafe.Pointer(&ty), d.r)
	if err != nil {
		return err
	}

	if ty != val.Type() {
		return encio.NewError(encio.ErrBadType, fmt.Sprintf("cannot set %v to received type %v", val.Type(), ty), 0)
	}

	enc := d.source.NewEncodable(ty, nil)
	return (*enc).Decode(unsafe.Pointer(val.UnsafeAddr()), d.r)
}
