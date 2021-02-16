package encodable

import (
	"errors"
	"fmt"
	"io"
	"reflect"
	"unsafe"

	"github.com/stewi1014/encs/encio"
)

// NewMultiAll returns a new MultiAll Encodable.
//
// It iterates through the slice of Encodables when Encoding or Decoding, returning the first error returned.
func NewMultiAll(encs []Encodable) MultiAll {
	if len(encs) > int(encio.MaxVarint) || len(encs) == 0 {
		panic(encio.NewError(
			errors.New("bad number of encodables"),
			fmt.Sprintf("cannot create MultiAll Encodable with %v encodables", len(encs)),
			0,
		))
	}

	t := encs[0].Type()
	for _, enc := range encs {
		if et := enc.Type(); et != t {
			panic(encio.NewError(
				encio.ErrBadType,
				fmt.Sprintf("Encodables for same type report different ones; have %v and %v types", t, et),
				0,
			))
		}
	}

	return MultiAll{
		encs: encs,
	}
}

// MultiAll provides methods for encoding using multiple Encodables.
//
// It allows multiple Encodables to be used for a single type, ensuring that all pass.
type MultiAll struct {
	encs []Encodable
}

// Size implements Encodable.
func (e *MultiAll) Size() int {
	size := 0
	for _, enc := range e.encs {
		size += enc.Size()
	}
	return size
}

// Type implements Encodable.
func (e *MultiAll) Type() reflect.Type {
	return e.encs[0].Type()
}

// Encode implements Encodable.
//
// It stops at and returns the first error.
func (e *MultiAll) Encode(ptr unsafe.Pointer, w io.Writer) error {
	for _, enc := range e.encs {
		if err := enc.Encode(ptr, w); err != nil {
			return err
		}
	}
	return nil
}

// Decode implements Encodable.
//
// It stops at and returns the first error.
func (e *MultiAll) Decode(ptr unsafe.Pointer, r io.Reader) error {
	for _, enc := range e.encs {
		if err := enc.Decode(ptr, r); err != nil {
			return err
		}
	}
	return nil
}

// NewMultiAny returns a new MultiAll Encodable.
//
// It iterates through the slice of Encodables when Encoding or Decoding, returning on the first successful call.
// It returns the first error if all fail.
func NewMultiAny(encs []Encodable) MultiAll {
	if len(encs) > int(encio.MaxVarint) || len(encs) == 0 {
		panic(encio.NewError(
			errors.New("bad number of encodables"),
			fmt.Sprintf("cannot create MultiAny Encodable with %v encodables", len(encs)),
			0,
		))
	}

	t := encs[0].Type()
	for _, enc := range encs {
		if et := enc.Type(); et != t {
			panic(encio.NewError(
				encio.ErrBadType,
				fmt.Sprintf("Encodables for same type report different ones; have %v and %v types", t, et),
				0,
			))
		}
	}

	return MultiAll{
		encs: encs,
	}
}

// MultiAny provides methods for encoding using multiple Encodables.
//
// It allows multiple Encodables to be used for a single type, returning on the first successful call.
type MultiAny struct {
	encs   []Encodable
	buff   encio.Buffer
	intEnc encio.Uint32
}

// Size implements Encodable.
func (e *MultiAny) Size() int {
	max := 0
	for _, enc := range e.encs {
		size := enc.Size()
		if size < 0 {
			return -1
		}

		if size > max {
			max = size
		}
	}

	return max + 4
}

// Type implements Encodable.
func (e *MultiAny) Type() reflect.Type {
	return e.encs[0].Type()
}

// Encode implements Encodable.
//
// It tries all Encodables, returning nil on the first successful encode.
// If no Encodables successfully encode, it returns the first error.
func (e *MultiAny) Encode(ptr unsafe.Pointer, w io.Writer) (err error) {
	for i, enc := range e.encs {
		e.buff.Reset()
		if encErr := enc.Encode(ptr, &e.buff); encErr != nil {
			if err == nil {
				err = encErr
			}
			continue
		}

		if err := e.intEnc.Encode(w, uint32(i)); err != nil {
			return err
		}

		return encio.Write(e.buff, w)
	}

	return
}

// Decode implements Encodable.
//
// It decodes with the Encodable at the same index as the encoding Encodable.
func (e *MultiAny) Decode(ptr unsafe.Pointer, r io.Reader) error {
	i, err := e.intEnc.Decode(r)
	if err != nil {
		return err
	}

	if i >= uint32(len(e.encs)) {
		return encio.NewError(
			encio.ErrMalformed,
			fmt.Sprintf("we have %v Encodables, but apparently we need to use Encodable at index %v to decode this", len(e.encs), i),
			0,
		)
	}

	return e.encs[i].Decode(ptr, r)
}
