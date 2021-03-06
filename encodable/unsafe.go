package encodable

import (
	"reflect"
	"unsafe"

	"github.com/stewi1014/encs/encio"
)

// checkPtr panics if ptr is nil.
// As per the documentation of unsafe, unsafe.Pointer types cannot be nil at any time. See notes in encodable.go.
func checkPtr(ptr unsafe.Pointer) {
	if ptr == nil {
		panic(encio.NewError(encio.ErrNilPointer, "unsafe.Pointer types are never allowed to be nil as per https://golang.org/pkg/unsafe/", 1))
	}
}

// newAt creates a new type of t, pointing ptr to it.
func newAt(ptr *unsafe.Pointer, t reflect.Type) {
	new := reflect.New(t)
	*ptr = unsafe.Pointer(new.Pointer())
}
