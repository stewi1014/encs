package encs

import (
	"io"
	"reflect"
	"sync"
	"unsafe"

	"github.com/stewi1014/encs/encio"
	"github.com/stewi1014/encs/encodable"
)

func NewEncoder(w io.Writer) *Encoder {
	return &Encoder{
		w:       w,
		typeEnc: encodable.NewType(0),
		source:  &encodable.DefaultSource{},
		encs:    make(map[reflect.Type]encodable.Encodable),
	}
}

type Encoder struct {
	w       io.Writer
	mutex   sync.Mutex
	typeEnc *encodable.Type
	source  encodable.Source
	encs    map[reflect.Type]encodable.Encodable
}

func (e *Encoder) Encode(v interface{}) error {
	if v == nil {
		return encio.NewError(encio.ErrNilPointer, "cannot encode nil interface", 0)
	}

	t := reflect.TypeOf(v)
	if t.Kind() != reflect.Ptr {
		return encio.NewError(encio.ErrBadType, "values must be passed by reference", 0)
	}

	e.mutex.Lock()
	defer e.mutex.Unlock()

	t = t.Elem() // the encoded type
	err := e.typeEnc.Encode(unsafe.Pointer(&t), e.w)
	if err != nil {
		return err
	}

	enc, ok := e.encs[t]
	if !ok {
		enc = e.source.NewEncodable(t, 0)
		e.encs[t] = enc
	}

	return enc.Encode(unsafe.Pointer(reflect.ValueOf(v).Elem().UnsafeAddr()), e.w)
}
