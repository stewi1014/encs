package encs

import "github.com/stewi1014/encs/encodable"

// I don't like how Config in this library and Config in enc are implemented slightly differently (Config here, *Config in enc),
// but the differences in internal usage warrant it. I'd like to use a non-pointer type in enc, but

// Config defines configuration for Encoders and Decoders
type Config struct {
	// Resolver is the encoder for reflect.Types.
	// If nil, the default resolver will be used, and Encoded types must be registered with encs.Register()
	Resolver encodable.Resolver

	//TODO: add more
}

func (c *Config) copyAndFill() *Config {
	config := new(Config)
	if c != nil {
		*config = *c
	}

	if config.Resolver == nil {
		config.Resolver = DefaultResolver
	}

	return config
}
