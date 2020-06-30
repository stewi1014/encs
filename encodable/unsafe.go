package encodable

import (
	"reflect"
	"unsafe"

	"github.com/stewi1014/encs/encio"
)

// ptr must be pointer to interface{}; not other interface types.
func ptrInterface(ptr unsafe.Pointer) *interfacePtr {
	return (*interfacePtr)(ptr)
}

type interfacePtr struct {
	typeInfo unsafe.Pointer
	elem     unsafe.Pointer // elem is not always what it seems. take care.
}

// for read only
func (i *interfacePtr) ptr() unsafe.Pointer {
	iv := reflect.NewAt(interfaceType, unsafe.Pointer(i)).Elem()
	if iv.Kind() == reflect.Ptr {
		return unsafe.Pointer(&i.elem)
	}
	return i.elem
}

func ptrSlice(ptr unsafe.Pointer) *slicePtr {
	return (*slicePtr)(ptr)
}

type slicePtr struct {
	array unsafe.Pointer
	len   int
	cap   int
}

// byteSliceAt returns a byteslice with the given length, using ptr as it's backing array.
func byteSliceAt(ptr uintptr, cap int) []byte {
	s := reflect.SliceHeader{
		Data: ptr,
		Len:  cap,
		Cap:  cap,
	}
	return *(*[]byte)(unsafe.Pointer(&s))
}

// newAt creates a new type of t, pointing ptr to it.
func newAt(ptr *unsafe.Pointer, t reflect.Type) {
	new := reflect.New(t)
	*ptr = unsafe.Pointer(new.Pointer())
}

// malloc allocates bytes bytes in the heap, pointing ptr to it.
func malloc(bytes uintptr, ptr *unsafe.Pointer) {
	buff := make([]byte, bytes)
	*ptr = *(*unsafe.Pointer)(unsafe.Pointer(&buff))
}

// checkPtr panics if ptr is nil.
// As per the documentation of unsafe, unsafe.Pointer types cannot be nil at any time. See notes in encodable.go.
func checkPtr(ptr unsafe.Pointer) {
	if ptr == nil {
		panic(encio.NewError(encio.ErrNilPointer, "unsafe.Pointer types are never allowed to be nil as per https://golang.org/pkg/unsafe/", 1))
	}
}
