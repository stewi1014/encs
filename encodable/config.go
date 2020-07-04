package encodable

import "reflect"

// Config contains settings and information for the generation of a new Encodable.
// Some Encodables do nothing with Config, and some require information from it.
//
// Config *must* be the same for Encoder and Decoder.
type Config struct {
	// Resolver is used by, and must be non-nil for, types which require type-resolution at Encode-Decode time.
	// That is, types which can reference new or unknown types (read: interface types).
	// This could conceivably differ from Encoder and Decoder, as long as the decoding Resolver can decode the encoding Resolvers output.
	// Feel free to experiment.
	Resolver Resolver

	// IncludeUnexported will include unexported struct fields in the encoded data.
	IncludeUnexported bool

	// If StructTag is set, only struct fields with the given tag will be encoded
	StructTag string
}

// String returns a string unique to the given configuration.
// Format is Config(options, StructTag: <StructTag>, Resolver: <Resolver>).
// Options are
// - u for IncludeUnexported
func (c *Config) String() string {
	// the main point here is to be concice over descriptive, speed is not of great concern either.
	// the string should uniquely represent the config, but should be as human-readable as is reasonable without cluttering the screen.
	// nil configs should format like &Config{}.
	// strings are for debugging and equality checks.

	if c == nil {
		return "Config()"
	}

	var elements = make([]string, 1)

	// options
	if c.IncludeUnexported {
		elements[0] += "u"
	}

	// other info

	if c.StructTag != "" {
		elements = append(elements, "StructTag: "+c.StructTag)
	}

	if c.Resolver != nil {
		elements = append(elements, "Resolver: "+Name(reflect.TypeOf(c.Resolver)))
	}

	str := "Config("
	if elements[0] != "" {
		str += " " + elements[0]
	}
	for i := 1; i < len(elements); i++ {
		if elements[i] != "" {
			str += ", " + elements[i]
		}
	}
	str += ")"
	return str
}

func (c *Config) genState() *state {
	s := &state{}

	if c != nil {
		s.Config = *c
	}

	return s
}

func (c *Config) copy() *Config {
	config := new(Config)
	if c != nil {
		*config = *c
	}
	return config
}

// state holds information about an Encodable, providing information about the Encodable tree to the elements of it.
// Encodables which have component Encodables should pass this on to their children on initialisation.
type state struct {
	Config

	// referencer is created by the first Encodable that asks for it.
	r *referencer
}

// callers *must* return Encodable as their new Encodable.
func (s *state) referencer(enc Encodable) (*referencer, Encodable) {
	if s.r == nil {
		s.r = &referencer{
			encoders: make(map[reflect.Type]*Concurrent),
			enc:      enc,
		}
		return s.r, s.r
	}
	return s.r, enc
}
