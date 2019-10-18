package enc

import (
	"io"
	"reflect"
	"sync"
	"unsafe"
)

// NewConcurrentEncodable returns a new concurrent-safe encodable.
func NewConcurrentEncodable(t reflect.Type, config *Config) *ConcurrentEncodable {
	if config != nil {
		config = config.copy()
	}
	return newConcurrentEncodable(t, config)
}

// retain config
func newConcurrentEncodable(t reflect.Type, config *Config) *ConcurrentEncodable {
	return &ConcurrentEncodable{
		t: t,
		c: config,
	}
}

// ConcurrentEncodable allows concurrent Encode and Decode operations on an Encodable.
type ConcurrentEncodable struct {
	t reflect.Type
	c *Config

	encodersMutex sync.Mutex
	encoders      []Encodable
}

// Size implements Sized
func (e *ConcurrentEncodable) Size() int {
	enc := e.get()
	defer e.put(enc)

	return enc.Size()
}

// Type implements Encodable
func (e *ConcurrentEncodable) Type() reflect.Type {
	return e.t
}

// Encode implements Encodable
func (e *ConcurrentEncodable) Encode(ptr unsafe.Pointer, w io.Writer) error {
	enc := e.get()
	defer e.put(enc)
	return enc.Encode(ptr, w)
}

// Decode implements Encodable
func (e *ConcurrentEncodable) Decode(ptr unsafe.Pointer, r io.Reader) error {
	enc := e.get()
	defer e.put(enc)
	return enc.Decode(ptr, r)
}

func (e *ConcurrentEncodable) get() Encodable {
	e.encodersMutex.Lock()
	defer e.encodersMutex.Unlock()
	l := len(e.encoders)
	if l > 0 {
		enc := e.encoders[l-1]
		e.encoders = e.encoders[:l-1]
		return enc
	}
	return newEncodable(e.t, e.c)
}

func (e *ConcurrentEncodable) put(enc Encodable) {
	e.encodersMutex.Lock()
	defer e.encodersMutex.Unlock()
	e.encoders = append(e.encoders, enc)
}
