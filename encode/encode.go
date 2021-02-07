package encode

import (
	"unsafe"

	"github.com/stewi1014/encs/encio"
)

// checkPtr panics if ptr is nil.
func checkPtr(ptr unsafe.Pointer) {
	if ptr == nil {
		panic(encio.NewError(encio.ErrNilPointer, "unsafe.Pointer types are never allowed to be nil as per https://golang.org/pkg/unsafe/", 1))
	}
}
