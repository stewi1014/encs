package encodable

import (
	"fmt"
	"reflect"
)

// Source is a generator of Encodables. Compound type Encodables take Source as an argument upon creation,
// and can use it for the generation of their element types either during creation or during encoding or decoding.
// Source is responsible for resolving recursive types and values, but this is not a requirement for implementing Source.
//
// There are a few implementations of Source in this library, and in many cases repeatedly wrapping Sources is helpful.
//
// For example, a Source may not implement any kind of handling for recursive types or values.
// If recursive types or values are involved, the Encodable returned by this Source may fail to accurately reproduce a pointer cycle,
// break assumptions Encodables must be able to make (e.g. no nested calls), or recurse infinitely at generation or encode-time.
// These kinds of 'dumb' Sources can define what encodables to use for what types, and then be wrapped with a generalised implementation of Source to provide more functionality.
//
// RecursiveSource for example, has no idea what Encodables should be used to encode a given type, rather, it can wrap a big switch statement and provide handling for resursive types.
type Source interface {
	// NewEncodable returns a new Encodable.
	// If non-nil, the Source given to this function is given to Encodables upon creation, otherwise it passes itself.
	//
	// It returns a pointer to an Encodable as it needs to be able to retroactively modify it. Subsequent Encodable generation could find that the type the Encodable
	// is encoding could be referenced, either during generation, or at encode-time through an interface or reflect.Value.
	// As such, in lieu of adding recursion avoidance/detection to all Encodables,
	// Source must be able to retroactively modify the encodable to wrap this functionality onto them, preferably with as little performance impact as possible.
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
	if enc == nil {
		panic(fmt.Sprintf("couldn't create encodable for %v", ty.String()))
	}
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
