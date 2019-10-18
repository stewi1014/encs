package gram

import (
	"unsafe"
)

func ptrSlice(ptr unsafe.Pointer) *slicePtr {
	return (*slicePtr)(ptr)
}

type slicePtr struct {
	array unsafe.Pointer
	len   int
	cap   int
}

func 

// Sliced returns the offset and true if reslice is a re-slice of b.
func Sliced(buff, reslice []byte) (off int, sliced bool) {
	index := *(*uintptr)(unsafe.Pointer(&reslice))
	begin := *(*uintptr)(unsafe.Pointer(&buff))

	off = int(index - begin)
	c := cap(buff)
	if off > c {
		return c, false
	}
	return int(off), true
}
