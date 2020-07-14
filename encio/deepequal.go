// +build !go1.15

package encio

// A bug in golang 1.14 causes DeepEqual to recurse infinitely. I was just about to make an issue for it before I found out it was fixed 27 days ago.
// In any case, it will be a while before it hits downstream, and a while after that before an acceptably small minority of people are using golang <=1.15.
// So, instead of disabling the test case that causes the recursion, we have the go 1.15 version of DeepEqual here with hacks to make it work outside of the reflect package.

import (
	"reflect"
	"unsafe"
)

const flagIndir uintptr = 1 << 7

var ptrOffset = func() *uintptr {
	field, ok := reflect.TypeOf(reflect.Value{}).FieldByName("ptr")
	if !ok {
		return nil
	}

	return &field.Offset
}()

var flagOffset = func() *uintptr {
	field, ok := reflect.TypeOf(reflect.Value{}).FieldByName("flag")
	if !ok {
		return nil
	}

	return &field.Offset
}()

func valueptr(v reflect.Value) unsafe.Pointer {
	if ptrOffset == nil {
		panic("ptr field of reflect.Value not found")
	}
	vptr := unsafe.Pointer(&v)
	return *(*unsafe.Pointer)(unsafe.Pointer(uintptr(vptr) + *ptrOffset))
}

func valueFlag(v reflect.Value) uintptr {
	if flagOffset == nil {
		panic("flag field of reflect.Value not found")
	}

	vptr := unsafe.Pointer(&v)
	return *(*uintptr)(unsafe.Pointer(uintptr(vptr) + *flagOffset))
}

func valuePointer(v reflect.Value) unsafe.Pointer {
	if valueFlag(v)&flagIndir != 0 {
		return *(*unsafe.Pointer)(valueptr(v))
	}
	return valueptr(v)
}

//go:linkname valueInterface reflect.valueInterface
func valueInterface(reflect.Value, bool) interface{}

// Copy of the DeepEqual function from go1.15, as it has infinite recursion fixed for recursive map types.
// Modified to work ourside the reflect package.
// See https://github.com/golang/go for licence.

// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Deep equality test via reflection

// During deepValueEqual, must keep track of checks that are
// in progress. The comparison algorithm assumes that all
// checks in progress are true when it reencounters them.
// Visited comparisons are stored in a map indexed by visit.
type visit struct {
	a1  unsafe.Pointer
	a2  unsafe.Pointer
	typ reflect.Type
}

// Tests for deep equality using reflected types. The map argument tracks
// comparisons that have already been seen, which allows short circuiting on
// recursive types.
func deepValueEqual(v1, v2 reflect.Value, visited map[visit]bool, depth int) bool {
	if !v1.IsValid() || !v2.IsValid() {
		return v1.IsValid() == v2.IsValid()
	}
	if v1.Type() != v2.Type() {
		return false
	}

	// We want to avoid putting more in the visited map than we need to.
	// For any possible reference cycle that might be encountered,
	// hard(v1, v2) needs to return true for at least one of the types in the cycle,
	// and it's safe and valid to get Value's internal pointer.
	hard := func(v1, v2 reflect.Value) bool {
		switch v1.Kind() {
		case reflect.Map, reflect.Slice, reflect.Ptr, reflect.Interface:
			// Nil pointers cannot be cyclic. Avoid putting them in the visited map.
			return !v1.IsNil() && !v2.IsNil()
		}
		return false
	}

	if hard(v1, v2) {
		// For a Ptr or Map value, we need to check flagIndir,
		// which we do by calling the pointer method.
		// For Slice or Interface, flagIndir is always set,
		// and using v.ptr suffices.
		ptrval := func(v reflect.Value) unsafe.Pointer {
			switch v.Kind() {
			case reflect.Ptr, reflect.Map:
				return valuePointer(v)
			default:
				return valueptr(v)
			}
		}
		addr1 := ptrval(v1)
		addr2 := ptrval(v2)
		if uintptr(addr1) > uintptr(addr2) {
			// Canonicalize order to reduce number of entries in visited.
			// Assumes non-moving garbage collector.
			addr1, addr2 = addr2, addr1
		}

		// Short circuit if references are already seen.
		typ := v1.Type()
		v := visit{addr1, addr2, typ}
		if visited[v] {
			return true
		}

		// Remember for later.
		visited[v] = true
	}

	switch v1.Kind() {
	case reflect.Array:
		for i := 0; i < v1.Len(); i++ {
			if !deepValueEqual(v1.Index(i), v2.Index(i), visited, depth+1) {
				return false
			}
		}
		return true
	case reflect.Slice:
		if v1.IsNil() != v2.IsNil() {
			return false
		}
		if v1.Len() != v2.Len() {
			return false
		}
		if v1.Pointer() == v2.Pointer() {
			return true
		}
		for i := 0; i < v1.Len(); i++ {
			if !deepValueEqual(v1.Index(i), v2.Index(i), visited, depth+1) {
				return false
			}
		}
		return true
	case reflect.Interface:
		if v1.IsNil() || v2.IsNil() {
			return v1.IsNil() == v2.IsNil()
		}
		return deepValueEqual(v1.Elem(), v2.Elem(), visited, depth+1)
	case reflect.Ptr:
		if v1.Pointer() == v2.Pointer() {
			return true
		}
		return deepValueEqual(v1.Elem(), v2.Elem(), visited, depth+1)
	case reflect.Struct:
		for i, n := 0, v1.NumField(); i < n; i++ {
			if !deepValueEqual(v1.Field(i), v2.Field(i), visited, depth+1) {
				return false
			}
		}
		return true
	case reflect.Map:
		if v1.IsNil() != v2.IsNil() {
			return false
		}
		if v1.Len() != v2.Len() {
			return false
		}
		if v1.Pointer() == v2.Pointer() {
			return true
		}
		for _, k := range v1.MapKeys() {
			val1 := v1.MapIndex(k)
			val2 := v2.MapIndex(k)
			if !val1.IsValid() || !val2.IsValid() || !deepValueEqual(val1, val2, visited, depth+1) {
				return false
			}
		}
		return true
	case reflect.Func:
		if v1.IsNil() && v2.IsNil() {
			return true
		}
		// Can't do better than this:
		return false
	default:
		// Normal equality suffices
		return valueInterface(v1, false) == valueInterface(v2, false)
	}
}

// DeepEqual doesn't recurse infinitely.
func DeepEqual(x, y interface{}) bool {
	if x == nil || y == nil {
		return x == y
	}
	v1 := reflect.ValueOf(x)
	v2 := reflect.ValueOf(y)
	if v1.Type() != v2.Type() {
		return false
	}
	return deepValueEqual(v1, v2, make(map[visit]bool), 0)
}
