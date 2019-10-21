package enc

import (
	"fmt"
	"io"
	"reflect"
	"sync"
	"unsafe"
)

// NewConcurrent returns a new concurrent-safe encodable.
func NewConcurrent(t reflect.Type, config *Config) *Concurrent {
	if config != nil {
		config = config.copy()
	}
	return newConcurrent(t, config)
}

// retain config
func newConcurrent(t reflect.Type, config *Config) *Concurrent {
	return &Concurrent{
		t: t,
		c: config,
	}
}

// Concurrent allows concurrent Encode and Decode operations on an Encodable.
type Concurrent struct {
	t reflect.Type
	c *Config

	encodersMutex sync.Mutex
	encoders      []Encodable
}

// Size implements Sized
func (e *Concurrent) Size() int {
	enc := e.get()
	defer e.put(enc)

	return enc.Size()
}

// Type implements Encodable
func (e *Concurrent) Type() reflect.Type {
	return e.t
}

func (e *Concurrent) String() string {
	enc := e.get()
	defer e.put(enc)
	return fmt.Sprintf("Concurrent(%v)", enc.String())
}

// Encode implements Encodable
func (e *Concurrent) Encode(ptr unsafe.Pointer, w io.Writer) error {
	enc := e.get()
	defer e.put(enc)
	return enc.Encode(ptr, w)
}

// Decode implements Encodable
func (e *Concurrent) Decode(ptr unsafe.Pointer, r io.Reader) error {
	enc := e.get()
	defer e.put(enc)
	return enc.Decode(ptr, r)
}

func (e *Concurrent) get() Encodable {
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

func (e *Concurrent) put(enc Encodable) {
	e.encodersMutex.Lock()
	defer e.encodersMutex.Unlock()
	e.encoders = append(e.encoders, enc)
}
