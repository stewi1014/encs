package main

import (
	"fmt"
	"reflect"
)

type recursiveStruct struct {
	A map[int]recursiveStruct
	B int
}

func main() {
	// Create struct.
	struct1 := recursiveStruct{
		A: make(map[int]recursiveStruct),
	}

	// Add struct to map.
	struct1.A[0] = struct1

	// Create second struct.
	struct2 := recursiveStruct{
		A: make(map[int]recursiveStruct),
	}

	// Add struct to map.
	struct2.A[0] = struct2

	fmt.Println(reflect.DeepEqual(struct1, struct2)) // Infinite recursion; stack overflow
}
