package encs

import "github.com/stewi1014/encs/enc"

var defaultTypeEncoder = enc.NewTypeRegistry(nil)

// Register registers the type of t. Types must be registered before encoding.
func Register(t interface{}) error {
	return defaultTypeEncoder.Register(t)
}
