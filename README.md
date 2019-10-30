# encs
A type-safe and modular encoding library  
[![GoDoc](https://godoc.org/github.com/stewi1014/encs?status.svg)](https://godoc.org/github.com/stewi1014/encs)

Encs aims to provide a type-strict, modular and feature-full encoding library with as little overhead as possible.

Goals include:
Type-safe: The type is encoded along with the value, and decoders will decode only into the same type that was sent, or in the case of interface encoding,
fill the interface with the same type as was sent. All types to be received must be Registered with Register()

Stream-promiscuious: Encoded messages are completely self-contained, and encoded streams can be picked up by a Decoder mid-stream and decoded sucessfully,
allowing a static Encoder to write to a dynamic number of receiving clients, and a dynamic number of sending clients to be decoded by a single Decoder.

Modular and Open: Methods for encoding are exposed in sub-packages, allowing their low-level encoding methods to be used to create custom encoding systems for a given use case,
without the overhead or added complexity of an Encoder or Decoder. The simple payload structure also allows easy re-implementation of the encs protocol.

encs/encodable provides encoders for specific types, and methods for encoding reflect.Type values.

encs/encio provides io and error types for encoding and related tasks

Example:
```go
	buff := new(bytes.Buffer)

	// This would typically go in an init() function.
	encs.Register(ExampleStruct{})

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

	err := enc.Encode(&example)
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
```