package encodable

import (
	"fmt"
	"io"
	"reflect"
	"sync"
	"unsafe"
)

// NewConcurrent wraps an Encodable with Concurrent, implementing Encodable with thread safety.
func NewConcurrent(newFunc func() Encodable) *Concurrent {
	return &Concurrent{
		new: newFunc,
	}
}

// Concurrent is a thread safe encodable.
// It functions as a drop in replacement for Encodables, keeping a cache of Encodables, only allowing a single call at a time on any one Encodable.
// If all cached Encodables are busy in a call, it creates a new Encodable, and calls it; It never blocks.
type Concurrent struct {
	new func() Encodable

	// Mutex is used to secure encoders. We must be careful never to hold the mutex in a way which might block calls,
	// that is, it must only be held for the moment when we modify encoders, and released before any other action.
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
	enc := e.get()
	defer e.put(enc)

	return enc.Type()
}

// String implements Encodable
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

// get returns a new Encodable, releasing ownership to the caller.
func (e *Concurrent) get() Encodable {
	e.encodersMutex.Lock()
	l := len(e.encoders)
	if l > 0 {
		enc := e.encoders[l-1]
		e.encoders = e.encoders[:l-1]
		e.encodersMutex.Unlock()
		return enc
	}
	e.encodersMutex.Unlock()
	return e.new()
}

// ownership of enc is passed to put, no more calls can be made.
func (e *Concurrent) put(enc Encodable) {
	e.encodersMutex.Lock()
	e.encoders = append(e.encoders, enc)
	e.encodersMutex.Unlock()
}
