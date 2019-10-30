package encodable

import (
	"reflect"
	"unsafe"
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

func ptrString(ptr unsafe.Pointer) *stringPtr {
	return (*stringPtr)(ptr)
}

type stringPtr struct {
	array unsafe.Pointer
	len   int
}

func (s *stringPtr) byteSlice() (buff []byte) {
	buffPtr := ptrSlice(unsafe.Pointer(&buff))
	buffPtr.array = s.array
	buffPtr.len = s.len
	buffPtr.cap = s.len
	return
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