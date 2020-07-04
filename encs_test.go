package encs_test

import (
	"bytes"
	"fmt"
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

	// This would typically go in an init() function.
	err := encs.Register(ExampleStruct{})
	if err != nil {
		panic(err)
	}

	birthday, _ := time.Parse(time.RFC3339, "2006-01-02T15:04:05Z")

	example := ExampleStruct{
		Name: "John Doe",
		Likes: []string{
			"Computers",
			"Music",
		},
		Birthday: birthday,
	}

	enc := encs.NewEncoder(buff, nil)

	err = enc.Encode(&example)
	if err != nil {
		fmt.Println(err)
		return
	}

	dec := encs.NewDecoder(buff, nil)

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
