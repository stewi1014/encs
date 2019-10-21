package enc

import "reflect"

// Config contains settings and information for the generation of a new Encodable.
// Some Encodables do nothing with Config, and some require information from it.
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
	// It's no possible to resolve these things within the scope of a single type's Encodable, so here it is.
	// implementation handled entirely internally.
	r *referencer
}

// String returns a string unique to the given configuration.
// Format is Config(options, tag: <StructTag>, te: <TypeEncoder>, resolver at <resolver location>).
// Options are
// - u or E for IncludeUnexported and not IncludeUnexported respectively.
func (c *Config) String() string {
	// the main point here is to be concice over descriptive.
	// the string should uniquely represent the config, but should be as human-readable as is reasonable,
	// without cluttering the screen. strings are for debugging.
	str := "Config("

	// options
	if c.IncludeUnexported {
		str += "u"
	} else {
		str += "E"
	}

	// other info

	if c.StructTag != "" {
		str += ", tag: " + c.StructTag
	}

	if c.Resolver != nil {
		str += ", te: " + reflect.TypeOf(c.Resolver).String()
	}

	if c.r != nil {
		str += ", resolver at " + c.r.Type().String()
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
