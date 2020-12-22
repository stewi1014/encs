package encs_test

import (
	"bytes"
	"fmt"
	"testing"
	"time"

	"github.com/stewi1014/encs"
)

type ExampleStruct struct {
	Name     string
	Likes    []string
	Birthday time.Time
}

func Example() {
	buff := new(bytes.Buffer)

	birthday, _ := time.Parse(time.RFC3339, "2006-01-02T15:04:05Z")

	example := ExampleStruct{
		Name: "John Doe",
		Likes: []string{
			"Computers",
			"Music",
		},
		Birthday: birthday,
	}

	enc := encs.NewEncoder(buff)

	err := enc.Encode(&example)
	if err != nil {
		fmt.Println(err)
		return
	}

	dec := encs.NewDecoder(buff)

	var decodedExample ExampleStruct
	err = dec.Decode(&decodedExample)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Printf("Name: %v, Likes: %v, Birthday: %v", decodedExample.Name, decodedExample.Likes, decodedExample.Birthday)

	// Output:
	// Name: John Doe, Likes: [Computers Music], Birthday: 2006-01-02 15:04:05 +0000 UTC
}

type ReferenceStruct1 struct {
	A *int
	B *int
}

func TestReferenceCycles(t *testing.T) {
	t.Run("References still share an object after decode", func(t *testing.T) {
		buff := new(bytes.Buffer)
		enc := encs.NewEncoder(buff)

		var i int = 1
		encode := ReferenceStruct1{
			A: &i,
			B: &i,
		}

		err := enc.Encode(&encode)
		if err != nil {
			t.Fatal(err)
		}

		dec := encs.NewDecoder(buff)
		var decode ReferenceStruct1
		err = dec.Decode(&decode)
		if err != nil {
			t.Fatal(err)
		}

		if decode.A == nil || decode.B == nil {
			t.Fatalf("Pointer is nil (%v, %v)", decode.A, decode.B)
		}

		if *decode.A != *encode.A || *decode.B != *encode.A {
			t.Fatalf("Wrong value. Wanted (%v, %v) but got (%v, %v)", *encode.A, *encode.B, *decode.A, *decode.B)
		}

		if decode.A != decode.B {
			t.Fatalf("Decoded pointers point to different addresses (%p and %p). Wanted the same address.", decode.A, decode.B)
		}
	})
}
