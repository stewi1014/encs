# [Encs](https://git.lenqua.net/stewi1014/encs/)
A featureful, fast, simple and modular encoding library
[![Go Reference](https://pkg.go.dev/badge/github.com/stewi1014/encs.svg)](https://pkg.go.dev/github.com/stewi1014/encs)
[![Go Report Card](https://goreportcard.com/badge/git.lenqua.net/stewi1014/encs)](https://goreportcard.com/report/git.lenqua.net/stewi1014/encs)
[![pipeline status](https://git.lenqua.net/stewi1014/encs/badges/master/pipeline.svg)](https://git.lenqua.net/stewi1014/encs/-/commits/master)
[![coverage report](https://git.lenqua.net/stewi1014/encs/badges/master/coverage.svg)](https://git.lenqua.net/stewi1014/encs/-/commits/master)

Encs provides methods for serialisation of golang types. 
A large part of the motivation of this library is that many encoding libraries seem to either lack features, be slow, hard to use, or have little opportunity for user expandability. If you don't need to serialise across languages, why not have all four?

Issues, Discussions and PRs are welcome.

# Goals
 * ## Simple to Use
	The default Encoder and Decoder provide out of the box functionality, and aims to provide a superset functionality of [golang/gob](https://golang.org/pkg/encoding/gob/). It can used in the same way, and aims to behave in the same way with the exception of extra features such as support for, and accurate reproduction of, recursive values, and encoding of types unsupported in gob; e.g. channels.

 * ## No Type is Out of Scope
	If it exists in the golang type system, encs aims to serialise it. Recursive types, Channels, reflect.Type and reflect.Value types are all encodable, and if you've got an io.ReadWriter, Functions too. Encs is tested agaist recursive values and types, and recreates pointer cycles accurately when decoding. Want to send a reflect.Value that is the value of itself? No worries.

 * ## Type Safe
	Encs will only decode into or create the exact same type as was sent by default. Encs provides the ability to accurately check type equality and compare the difference between types, and allows users to decide how type equivelency should be evaluated.

 * ## Stream Promiscuous
	Streams have no state; each encoded value is completely independent, and decodable without any extra information. Streams from encoders can be picked up by decoders mid-stream and decoded successfully, allowing a single Encoder to write to a dynamic number of receiving Decoders, and a dynamic number of sending Encoders to be decoded by a single Decoder, or both at the same time.

 * ## Modular
	Each golang type has an implementation of the encs/encode.Encodable interface dedicated to encoding the type. The serialisation logic, an Encodable, can be written and integrated into encs without worrying about type safety, encoding sub-types or recusing infinitely over a recursive type or value. This is handled by encs/encode.Source, which again, is self contained, extendable, logically divided and swappable with other implementations. I encorage you to have a look at the Encoder and Decoder. You'll find it's just gluing together various modules to provide functionality simmilar to [golang/gob](https://golang.org/pkg/encoding/gob/).

 * ## Open
	Methods for encoding are exposed in sub-packages, allowing the lower-level encoding methods to be used directly. If you have a look at the default encs Encoder/Decoder, you'll find it's just gluing together a few modules for typical use cases. If you just have a single struct type you want to send, you can skip the shenanigans and just use encs/encode.NewStruct().

 * ## Tested
	Encs is tested against test cases designed to make it fail, and PRs with test cases are so incredibly welcome.
 	Due to the contrived test values (i.e. a map with itself as an element), it is difficult to generalise testing, with even the standard library reflect.DeepEqual in go 1.14 recursing infinitely in some cases. Big thanks to [maxatome](https://github.com/maxatome) who has been incerdibly helpful with [go-testdeep](https://github.com/maxatome/go-testdeep) in supporting these cases.

# Non Goals
 * ## Standardised Format
	The serealised output of Encs is not standardised and is dependent on almost every component part of the encoder. Decoders must be constructed in exactly the same way as the encoder in order to sucessfully decode.
	Methods exist for ensuring this and it is not an issue for the default encoder/decoder within the same major version, as regression testing is performed on them.
	However, customising Encs will require you to do your own testing if you want garuntee compatilility between versions of your software.

 * ## Cross-Language Serialisation
	If you want to try to use encs to serialise to/from a different language, you're welcome to try, but I won't lie and say it will be anything but a massive pain and waste of time.
	Use something else; Encs is heavily go-centric.

 * ## Encoding Non-Standard Values
	If you use unsafe.Pointer to create an instance of *int64 that points to the last 32bits of one struct and the first 32bits of an adjacent struct,
	don't expact that kind of reference to be encoded while retaining the same reference structure.

	As a general rule, if a value is only reachable through skirting the golang type system Encs makes no garuntee to encode it.

# Use Cases
 * ## Process to Process Communication
	A daemon that may have multiple processes can use Encs to serialise large parts of its state. Heck, you could put your whole program in a "main" struct that encs will happily encode,
	reference cycles and all, allowing a new process to continue operating from exactly same state as the other process.
 * ## Distributed Processing
	Encs aims to act as a value "portal" for golang values.
	A future goal is to integreate Encs with an Entity-Component-System library to provide a framework for multiple processes to function on the same syncronised state,
	and this kind of usage is exactly what Encs aims to support.
# Packages

## encs/encodable
Contains the low-level encoding logic for golang types.
It provides encoders for golang types, and methods for resolving recursive types and values.

## encs/encio
Provides io methods and error types for encoding and related tasks, including multiplexing and functions for encoding integers.

# Usage
If you've used [golang/gob](https://golang.org/pkg/encoding/gob/) before, you should have no issue. They aim to behave in the same way as gob, but will successfully encode things that gob won't.

```go
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
```
