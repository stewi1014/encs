package encodable

import (
	"errors"
	"fmt"
	"hash"
	"hash/crc64"
	"io"
	"reflect"
	"sync"

	"github.com/stewi1014/encs/encio"
)

// Resolver is a method for encoding types.
type Resolver interface {
	// Encode encodes the type t, writing to w.
	Encode(t reflect.Type, w io.Writer) error

	// Decode reads an encoded type, returning it.
	// Decode must only read the same number of bytes written by Encode().
	//
	// In the case of decoding into interfaces, it might be expected that the received type is the same as the existing type in the interface.
	// in this case, it might be sufficient to simply check if the types are equal, relying on the expected type as an existing reflect.Type instance.
	// In this case, bool or (bool, error) would be sufficient return values.
	//
	// Other implementations might attempt to completely resolve a reflect.Type value from encoded data,
	// in which case expected is an uneccecary argument.
	//
	// In the case of registration-based Resolvers, if the interface contains the type that's being sent,
	// it can be fortuitus to register the expected type, as it might not have been registered before.
	//
	// I believe this function format is a good happy medium to allow for these different implementations,
	// however, Resolvers should always be certain that the type returned from decode is either the same as what was given to Encode, or nil.
	// Incorrect types will cause panics, or worse, incorrect memory manipulation.
	Decode(expected reflect.Type, r io.Reader) (reflect.Type, error)

	// Size returns the number of bytes the Resolver will read and write to the buffer.
	// It can write less, but never more. If length is undefined, Size should return a negative.
	Size() int
}

var (
	// ErrAlreadyRegistered is returned if a type is already registered.
	ErrAlreadyRegistered = errors.New("already registered")

	// ErrNotRegistered is returned if a type has not been registered.
	ErrNotRegistered = errors.New("not registered")
)

// NewRegisterResolver returns a new RegisterResolver TypeResolver
func NewRegisterResolver(hasher hash.Hash64) *RegisterResolver {
	if hasher == nil {
		hasher = crc64.New(crc64.MakeTable(crc64.ISO))
	}

	rr := &RegisterResolver{
		hasher:   hasher,
		idByType: make(map[reflect.Type][8]byte),
		typeByID: make(map[[8]byte]reflect.Type),
	}

	for _, T := range builtin {
		if err := rr.Register(T); err != nil {
			panic(err)
		}
	}

	return rr
}

// RegisterResolver is a registration-based TypeResolver.
// All types to be encoded and decoded must be registered with Register(),
// with the exception of int*, uint*, float*, complex*, string, bool, time.Time, and time.Duration, which are pre-registered.
// It is thread safe.
type RegisterResolver struct {
	hasher      hash.Hash64
	hasherMutex sync.Mutex

	idByType map[reflect.Type][8]byte
	typeByID map[[8]byte]reflect.Type
	mapMutex sync.Mutex
}

// Register registers T, &T, []T, and *T if T is a pointer.
func (rr *RegisterResolver) Register(t interface{}) error {
	var ty reflect.Type
	var ok bool
	if ty, ok = t.(reflect.Type); !ok {
		ty = reflect.TypeOf(t)
	}

	if err := rr.hashAndPut(ty); err != nil {
		return err
	}

	st := reflect.SliceOf(ty)
	if err := rr.hashAndPut(st); err != nil && err != ErrAlreadyRegistered {
		return err
	}

	pt := reflect.PtrTo(ty)
	if err := rr.hashAndPut(pt); err != nil && err != ErrAlreadyRegistered {
		return err
	}

	if ty.Kind() == reflect.Ptr {
		if err := rr.hashAndPut(ty.Elem()); err != nil && err != ErrAlreadyRegistered {
			return err
		}
	}

	return nil
}

func (rr *RegisterResolver) hashAndPut(ty reflect.Type) error {
	h, err := rr.hash(ty)
	if err != nil {
		return err
	}
	return rr.put(ty, h)
}

func (rr *RegisterResolver) hash(ty reflect.Type) (out [8]byte, err error) {
	rr.hasherMutex.Lock()
	defer rr.hasherMutex.Unlock()
	rr.hasher.Reset()
	buff := []byte(Name(ty))
	n, err := rr.hasher.Write(buff)
	if err != nil {
		return out, encio.NewError(err, "hash error", 0)
	}
	if n != len(buff) {
		return out, encio.NewError(io.ErrShortWrite, fmt.Sprintf("wrote %v, want %v", n, len(buff)), 0)
	}
	h := rr.hasher.Sum64()
	out[0] = uint8(h)
	out[1] = uint8(h >> 8)
	out[2] = uint8(h >> 16)
	out[3] = uint8(h >> 24)
	out[4] = uint8(h >> 32)
	out[5] = uint8(h >> 40)
	out[6] = uint8(h >> 48)
	out[7] = uint8(h >> 56)
	return
}

func (rr *RegisterResolver) put(ty reflect.Type, h [8]byte) error {
	rr.mapMutex.Lock()
	defer rr.mapMutex.Unlock()
	if oty, ok := rr.typeByID[h]; ok {
		if oty == ty {
			return encio.NewError(ErrAlreadyRegistered, fmt.Sprintf("type %v", ty), 1)
		}
		return encio.NewError(ErrAlreadyRegistered, fmt.Sprintf("hash of %v and %v are both %v", ty, oty, h), 1)
	}
	rr.typeByID[h] = ty
	rr.idByType[ty] = h
	return nil
}

func (rr *RegisterResolver) getByType(ty reflect.Type) ([8]byte, bool) {
	rr.mapMutex.Lock()
	h, ok := rr.idByType[ty]
	rr.mapMutex.Unlock()
	return h, ok
}

func (rr *RegisterResolver) getByID(h [8]byte) (reflect.Type, bool) {
	rr.mapMutex.Lock()
	ty, ok := rr.typeByID[h]
	rr.mapMutex.Unlock()
	return ty, ok
}

// Size implements TypeResolver
func (rr *RegisterResolver) Size() int {
	return 8
}

// Encode implements TypeResolver
func (rr *RegisterResolver) Encode(ty reflect.Type, w io.Writer) error {
	if h, ok := rr.getByType(ty); ok {
		return encio.Write(h[:], w)
	}

	// ty is not registered.
	// register it now and encode, returning an ErrNotRegistered error.
	h, err := rr.hash(ty)
	if err != nil {
		return encio.NewError(ErrNotRegistered, fmt.Sprintf("%v not previously registered. registering now failed with %v", ty, err), 0)
	}

	err = rr.put(ty, h)
	if err != nil {
		return encio.NewError(ErrNotRegistered, fmt.Sprintf("%v not previously registered. registering now failed with %v", ty, err), 0)
	}

	err = encio.Write(h[:], w)
	if err != nil {
		return encio.NewError(ErrNotRegistered, fmt.Sprintf("%v not previously registered but registered now. writing failed with %v", ty, err), 0)
	}
	return encio.NewError(ErrNotRegistered, fmt.Sprintf("%v written but not previously registered. decoder may not understand it", ty), 0)
}

// Decode implements TypeResolver
func (rr *RegisterResolver) Decode(expected reflect.Type, r io.Reader) (reflect.Type, error) {
	var h [8]byte
	if err := encio.Read(h[:], r); err != nil {
		return nil, err
	}

	if ty, ok := rr.getByID(h); ok {
		return ty, nil
	}

	if expected == nil {
		return nil, encio.NewError(ErrNotRegistered, fmt.Sprintf("received hash %v doesn't map to any known types. Is it registered?", h), 0)
	}

	eh, err := rr.hash(expected)
	if err != nil {
		return nil, encio.NewError(err, "couldn't hash expected type", 0)
	}
	if eh != h {
		return nil, encio.NewError(ErrNotRegistered, fmt.Sprintf("received hash %v doesn't map to any known types or the expected type. Is it registered?", h), 0)
	}

	// nolint
	rr.put(expected, eh)
	// Ignore errors from put; we've already succeeded in the decode (through the expected),
	// and if it really does need to be registered now then ErrNotRegistered will be returned by a later call.

	return expected, nil
}
