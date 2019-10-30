package encs

import "github.com/stewi1014/encs/encodable"

// DefaultResolver is the default Resolver used when Config.Resolver is nil.
// Types must be registered with Register()
var DefaultResolver = encodable.NewRegisterResolver(nil)

// Register registers the type of t. Types must be registered before encoding.
// It is a shortcut for DefaultResolver.Register()
func Register(t interface{}) error {
	return DefaultResolver.Register(t)
}
