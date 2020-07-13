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
		source:  encodable.NewCachingSource(encodable.NewRecursiveSource(encodable.New)),
	}
}

type Encoder struct {
	w       io.Writer
	mutex   sync.Mutex
	typeEnc *encodable.Type
	source  encodable.Source
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

	return e.source.NewEncodable(t, 0).Encode(unsafe.Pointer(reflect.ValueOf(v).Elem().UnsafeAddr()), e.w)
}
