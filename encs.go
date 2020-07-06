// Package encs aims to provide a type-strict, modular and feature-full encoding library with as little overhead as possible.
//
// Goals include:
// Type-safe: The type is encoded along with the value, and decoders will decode only into the same type that was sent, or in the case of interface encoding,
// fill the interface with the same type as was sent. All types to be received must be Registered with Register()
//
// Stream-promiscuous: Encoded messages are completely self-contained, and encoded streams can be picked up by a Decoder mid-stream and decoded successfully,
// allowing a static Encoder to write to a dynamic number of receiving clients, and a dynamic number of sending clients to be decoded by a single Decoder.
//
// Modular and Open: Methods for encoding are exposed in sub-packages, allowing their low-level encoding methods to be used to create custom encoding systems for a given use case,
// without the overhead or added complexity of an Encoder or Decoder. The simple payload structure also allows easy re-implementation of the encs protocol.
//
// encs/encodable provides encoders for specific types, and methods for encoding reflect.Type values.
//
// encs/encio provides io and error types for encoding and related tasks
package encs

import (
	"reflect"

	"github.com/stewi1014/encs/encodable"
)

// Register registers a type to be encoded.
func Register(types ...interface{}) error {
	rtypes := make([]reflect.Type, len(types))

	for i := range types {
		rtypes[i] = reflect.TypeOf(types)
	}

	return encodable.Register(rtypes...)
}
