package encs

import (
	"io"
	"reflect"
	"sync"
	"unsafe"

	"github.com/stewi1014/encs/encio"
	"github.com/stewi1014/encs/encodable"
	"github.com/stewi1014/encs/encode"
)

func NewEncoder(w io.Writer) *Encoder {
	return &Encoder{
		w:       w,
		typeEnc: encode.NewType(false),
		source:  encodable.NewCachingSource(encode.NewRecursiveSource(DefaultSource)),
	}
}

type Encoder struct {
	w       io.Writer
	mutex   sync.Mutex
	typeEnc *encode.Type
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

	enc := e.source.NewEncodable(t, nil)
	return (*enc).Encode(unsafe.Pointer(reflect.ValueOf(v).Elem().UnsafeAddr()), e.w)
}
