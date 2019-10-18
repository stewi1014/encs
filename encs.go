// Package gneg stands for Gneg does Not Encode Generics; a Type-Strict encoding library.
//
// Gneg provides a serilisation method for golang types. It aims to be configurable and portable.
// Features include:
//
// Stream Promiscuity*; encoded streams can be picked up by a decoder mid-stream and decoded sucessfully,
// allowing a single encoder to write to a dynamic number of receiving clients, and a dynamic number of sending clients to be
// decoded by a single decoder.
//
// Type-safe; Types are sent with their encoded value (See TypeResolver), allowing decoders to create the sent type locally and decode into it.
// The type returned by Decode (+pointer) is always the same as the type given to Encode.
//
// Notes on Decoding:
// When decoding into an interface, it must be passed by reference (Decode(&val)) unless PreserveInterface is set, and the interface points to the same type as
// the type passed to Encode. If PreserveInterface is set, the decoding type must match the Encoded type.
//
// The sub-package gram provides helpful functions for creating and reading messages.
package encs
