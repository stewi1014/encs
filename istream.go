// Package gneg stands for Gneg does Not Encode Generics.
//
// Gneg provides a serilisation method for golang types. It aims to be modular, configurable and portable.
// Features include:
//
// Stream Promiscuity; encoded streams can be picked up by a decoder at any point and decoded sucessfully,
// allowing a single encoder to write to a dynamic number of receiving clients, and a dynamic number of sending clients to be
// decoded by a single decoder.
//
// Interface supporting; types are encoded (with different configurations) with the sent data, allowing decoders to decode into interfaces.
//
// Modularity; encoders and decoders use a TypeResolver to encode and decode types, and gram.Reader and gram.Writer to write compiled messages to the wire.
// These can be replaced with any system implementing these interfaces, allowing custom type encoding and payload reading/writing systems to be used.
//
// The sub-package gram provides helpful functions for creating and reading messages.
package gneg

import (
	"encoding/binary"
	"errors"

	"github.com/stewi1014/gneg/gram"
)

const (
	// 4 or 8 bytes for an int.
	wordSize = 4 << ((^uint(0) >> 32) & 1)

	// 16gb on 64bit systems, 1gb on 32.
	tooBig = 1 << (wordSize + 26)
)

var (
	errTooBig = errors.New("buffer overflow")
	binEnc    = binary.LittleEndian
)

// Config contains configuration for how streams are encoded and decoded.
// Nil values are default.
type Config struct {
	// TypeResolver is the type encoding system used by the Encoder and Decoder.
	// If nil, the default internal RegisterResolver will be used, and types should be registered with
	// the package-level function Register().
	TypeResolver TypeResolver

	// GramEncoder is a system for writing grams onto a wire.
	GramEncoder gram.Writer
	GramDecoder gram.Reader
}
