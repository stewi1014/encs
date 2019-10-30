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

/*

// Config contains settings and information for the generation of a new Encodable.
// Some Encodables do nothing with Config, and some require information from it.
//
// Config also holds generated configuration pertaining to a specific Encodable,
// so to avoid breaking previously created Encodables with subsequent changes to Config,
// the New* functions of Encodables copy the config.
//
// Config must be the same for Encoder and Decoder.
type Config struct {
	// Resolver is used by, and must be non-nil for, types which require type-resolution at Encode-Decode time.
	// That is, types which can reference new or unknown types (read: interface types).
	Resolver Resolver

	// IncludeUnexported will include unexported struct fields in the encoded data.
	IncludeUnexported bool

	// If StructTag is set, only struct fields with the given tag will be encoded
	StructTag string

	// used by pointer-types to resolve references;
	// i.e. multiple pointers to the same value, recursive references .. should all retain their reference structure.
	// It's not possible to resolve these things within the scope of a single type's Encodable, so here it is.
	// implementation handled internally.
	r *referencer
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

	// This should never be non-nil for caller Config values; it is only set internally,
	// however, it is still important for equality checks.
	if c.r != nil {
		elements = append(elements, "ref resolver at "+Name(c.r.Type()))
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

// callers *must* return Encodable as their new Encodable.
func (c *Config) referencer(enc Encodable) (*referencer, Encodable) {
	if c.r == nil {
		c.r = &referencer{
			encoders: make(map[reflect.Type]*Concurrent),
			enc:      enc,
		}
		return c.r, c.r
	}
	return c.r, enc
}

func (c *Config) copy() *Config {
	config := new(Config)
	*config = *c
	return config
}

*/
