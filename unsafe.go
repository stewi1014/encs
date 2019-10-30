package encs

import "unsafe"

// ptr must be pointer to interface{}; not other interface types.
func ptrInterface(ptr unsafe.Pointer) *interfacePtr {
	return (*interfacePtr)(ptr)
}

type interfacePtr struct {
	typeInfo unsafe.Pointer
	elem     unsafe.Pointer // elem is not always what it seems. take care.
}
