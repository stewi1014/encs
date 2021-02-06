package encode

import (
	"reflect"
)

// Source is a generator of Encodables. Compound type Encodables take Source as an argument upon creation,
// and can use it for the generation of their element types either during creation or during encoding or decoding.
//
// Source is responsible for resolving recursive types and values if needed, both during generation and at encode/decode time,
// but this is not a requirement for implementing Source; it's simply a feature wanted in most use cases.
// A simple switch statement over types and generating Encodables for them can suffice.
//
// There are a few implementations of Source in this library, and in many cases repeatedly wrapping Sources that provide different features is helpful.
// RecursiveSource for example, has no idea what Encodables should be used to encode a given type, rather, it can wrap a Source which does, and add handling for resursive types and values.
type Source interface {
	// NewEncodable returns a new Encodable to be used to serialise the given type.
	//
	// It returns a pointer to an Encodable as it needs to be able to retroactively modify it. Subsequent Encodable generation could find that the type the Encodable
	// is encoding could be referenced, either through a static recursive type, or at encode-time through a type like interface{} or reflect.Value.
	// As such, in lieu of adding recursion avoidance/detection to all Encodables,
	// Source must be able to retroactively modify the encodable to wrap this functionality onto it.
	//
	// The Source passed to NewEncodable must be passed to the Encodable that it creates. It is used by wrapping Sources to pass themselves to new Encodables,
	// so they don't loose control of element Encodable generation.
	NewEncodable(reflect.Type, Source) *Encodable
}

// SourceFromFunc creates a source from a function.
// It will substitute itself if NewEncodable() is called with a nil-source.
//
// It will also panic if nil is returned.
func SourceFromFunc(source func(reflect.Type, Source) Encodable) Source {
	return funcSource{newEncodable: source}
}

type funcSource struct {
	newEncodable func(reflect.Type, Source) Encodable
}

func (s funcSource) NewEncodable(ty reflect.Type, source Source) *Encodable {
	if source == nil {
		source = s
	}
	enc := s.newEncodable(ty, source)
	return &enc
}

// NewCachingSource returns a new CachingSource, using source for cache misses.
// Users of CachingSource must not pass it to element Encodables who may try to create themselves,
// else a situation may arise where a recursive type causes an Encodable to be given itself from the cache, and it makes
// nested calls to itself. Use the original, CachingSource.Source to pass to element Encodables, assuming it properly resolves recursive types.
func NewCachingSource(source Source) *CachingSource {
	return &CachingSource{
		cache:  make(map[reflect.Type]*Encodable),
		Source: source,
	}
}

// CachingSource provides a cache of Encodables.
type CachingSource struct {
	cache map[reflect.Type]*Encodable
	Source
}

// NewEncodable implements Source.
func (src *CachingSource) NewEncodable(ty reflect.Type, parent Source) (enc *Encodable) {
	enc, ok := src.cache[ty]
	if ok {
		return enc
	}

	enc = src.Source.NewEncodable(ty, parent)
	src.cache[ty] = enc
	return enc
}
