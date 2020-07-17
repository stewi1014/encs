package encodable

import (
	"reflect"
)

// EncID is comparable struct representing an encodable of a particular type and configuration.
// It can be used as keys in maps of Encodables to confirm equality across Encodables.
type EncID struct {
	reflect.Type
	Config
}

// Source is a generator of Encodables. Compound type Encodables take Source as an argument upon creation,
// and can use it for the generation of their element types either during creation or during encoding or decoding.
// **The Source is responsible for resolving recursive types and values**.
type Source interface {
	// NewEncodable returns a new Encodable.
	// Source should take care to avoid infinite recursion, taking note of when it is called to create an Encodable from inside the same Encodable's creation function,
	// or when a type could possibly reference a parent encodable's value. The Encodable may not be de-referenced; Source may retroactively replace the Encodable.
	//
	// Implementations of Source can use Recursive as an Encodable that solves recursive cases.
	//
	// The Source passed to NewEncodable is passed to the Encodable that it creates. It is used by wrapping Sources to pass themselves to new Encodables,
	// so they don't loose control of element Encodable generation.
	NewEncodable(reflect.Type, Config, Source) *Encodable
}

// SourceFromFunc creates a source using a function. It is mostly used for spoofing tests, but can do other things.
// Typically, Sources need to retain information between Encodable generation for various reasons, including avoiding infinite recursion
// and caching for speed. Usability is typically limited to testing and creating "simple" sources that are wrapped with other sources providing neccecary features.
func SourceFromFunc(newEncodable func(reflect.Type, Config, Source) *Encodable) Source {
	return funcSource{newEncodable: newEncodable}
}

type funcSource struct {
	newEncodable func(reflect.Type, Config, Source) *Encodable
}

func (s funcSource) NewEncodable(ty reflect.Type, config Config, source Source) *Encodable {
	return s.newEncodable(ty, config, source)
}

// NewCachingSource returns a new CachingSource, using source for cache misses.
// Users of CachingSource must not pass it to element Encodables who may try to create themselves,
// else a situation may arise where a recursive type causes an Encodable to be given itself from the cache, and it makes
// nested calls to itself. Use the original, CachingSource.Source to pass to element Encodables, assuming it properly resolves recursive types.
func NewCachingSource(source Source) *CachingSource {
	return &CachingSource{
		cache:  make(map[EncID]*Encodable),
		Source: source,
	}
}

// CachingSource provides a cache of Encodables.
type CachingSource struct {
	cache map[EncID]*Encodable
	Source
}

// NewEncodable implements Source.
func (src *CachingSource) NewEncodable(ty reflect.Type, config Config, parent Source) (enc *Encodable) {
	enc, ok := src.cache[EncID{Type: ty, Config: config}]
	if ok {
		return enc
	}

	enc = src.Source.NewEncodable(ty, config, parent)
	src.cache[EncID{Type: ty, Config: config}] = enc
	return enc
}
