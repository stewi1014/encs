package encodable

import (
	"errors"
	"fmt"
	"hash/crc64"
	"reflect"
	"strconv"
	"time"

	"github.com/stewi1014/encs/encio"
)

var (
	reflectTypeType  = reflect.TypeOf(new(reflect.Type)).Elem()
	reflectValueType = reflect.TypeOf(new(reflect.Value)).Elem()
)

func init() {
	err := Register(
		nil,
		reflect.TypeOf(int(0)),
		reflect.TypeOf(int8(0)),
		reflect.TypeOf(int16(0)),
		reflect.TypeOf(int32(0)),
		reflect.TypeOf(int64(0)),
		reflect.TypeOf(uint(0)),
		reflect.TypeOf(uint8(0)),
		reflect.TypeOf(uint16(0)),
		reflect.TypeOf(uint32(0)),
		reflect.TypeOf(uint64(0)),
		reflect.TypeOf(uintptr(0)),
		reflect.TypeOf(float32(0)),
		reflect.TypeOf(float64(0)),
		reflect.TypeOf(complex64(0)),
		reflect.TypeOf(complex128(0)),
		reflect.TypeOf(string("")),
		reflect.TypeOf(bool(false)),
		reflect.TypeOf(time.Time{}),
		reflect.TypeOf(time.Duration(0)),
		reflectTypeType,
		reflectValueType,
	)

	if err != nil {
		panic(err)
	}
}

var registered []reflect.Type

// ErrAlreadyRegistered is returned by Register if one or more of the given types has already been registered.
// It is wrapped.
var ErrAlreadyRegistered = errors.New("already registered")

// Register allows an unknown type to be decoded with Type.
func Register(types ...reflect.Type) error {
	var errMsg string
	for _, ty := range types {
		if !register(ty) {
			if len(errMsg) > 0 {
				errMsg += ", "
			}
			errMsg += fmt.Sprintf("%v", ty)
			continue
		}
	}

	if errMsg != "" {
		return fmt.Errorf("%w: %v", ErrAlreadyRegistered, errMsg)
	}
	return nil
}

// register registers the type, returning false if it was previously registered.
func register(ty reflect.Type) bool {
	for _, r := range registered {
		if r == ty {
			return false
		}
	}
	registered = append(registered, ty)
	return true
}

// GetID returns a unique ID for a given reflect.Type.
func GetID(ty reflect.Type, config Config) (id ID) {
	if ty == nil {
		return ID{0, 0}
	}

	hasher := crc64.New(crc64.MakeTable(crc64.ISO))

	// First half is the loose glob.
	if err := encio.Write([]byte(looseGlob(ty, nil)), hasher); err != nil {
		panic(err)
	}

	id[0] = hasher.Sum64()

	// Second half is the details about the types kind
	hasher.Reset()
	if err := encio.Write([]byte(strictGlob(ty, nil)), hasher); err != nil {
		panic(err)
	}

	id[1] = hasher.Sum64()

	return id
}

// ID is a 128 bit ID for a given reflect.Type.
type ID [2]uint64

func (id ID) String() string {
	return fmt.Sprintf("%#x%x", id[0], id[1])
}

// Encode writes the ID to the first 16 bytes in buff.
// It panics if len(buff) < 16.
func (id *ID) Encode(buff []byte) {
	buff[0] = uint8(id[0])
	buff[1] = uint8(id[0] >> 8)
	buff[2] = uint8(id[0] >> 16)
	buff[3] = uint8(id[0] >> 24)
	buff[4] = uint8(id[0] >> 32)
	buff[5] = uint8(id[0] >> 40)
	buff[6] = uint8(id[0] >> 48)
	buff[7] = uint8(id[0] >> 56)
	buff[8] = uint8(id[1])
	buff[9] = uint8(id[1] >> 8)
	buff[10] = uint8(id[1] >> 16)
	buff[11] = uint8(id[1] >> 24)
	buff[12] = uint8(id[1] >> 32)
	buff[13] = uint8(id[1] >> 40)
	buff[14] = uint8(id[1] >> 48)
	buff[15] = uint8(id[1] >> 56)
}

// Decode reads the ID from the first 16 bytes in buff.
// It panics if len(buff) < 16.
func (id *ID) Decode(buff []byte) {
	id[0] = uint64(buff[0])
	id[0] |= uint64(buff[1]) << 8
	id[0] |= uint64(buff[2]) << 16
	id[0] |= uint64(buff[3]) << 24
	id[0] |= uint64(buff[4]) << 32
	id[0] |= uint64(buff[5]) << 40
	id[0] |= uint64(buff[6]) << 48
	id[0] |= uint64(buff[7]) << 56
	id[1] = uint64(buff[8])
	id[1] |= uint64(buff[9]) << 8
	id[1] |= uint64(buff[10]) << 16
	id[1] |= uint64(buff[11]) << 24
	id[1] |= uint64(buff[12]) << 32
	id[1] |= uint64(buff[13]) << 40
	id[1] |= uint64(buff[14]) << 48
	id[1] |= uint64(buff[15]) << 56
}

// looseGlob returns a glob of data which should loosely identify types.
// int* uint* and float* are considered equal.
// complex64 and complex128 are considered equal.
// structs are considered equal if they have the same name.
// interfaces are considered equal if the have the same name.
func looseGlob(ty reflect.Type, seen map[reflect.Type]int) (str string) {
	if seen == nil {
		seen = make(map[reflect.Type]int)
	}

	switch ty.Kind() {
	case reflect.Invalid:
		str += "invalid"
		return

	case reflect.Bool:
		str += "bool"
		return

	case reflect.Int,
		reflect.Int8,
		reflect.Int16,
		reflect.Int32,
		reflect.Int64,
		reflect.Uint,
		reflect.Uint8,
		reflect.Uint16,
		reflect.Uint32,
		reflect.Uint64,
		reflect.Uintptr,
		reflect.Float32,
		reflect.Float64:
		str += "number"
		return

	case reflect.Complex64,
		reflect.Complex128:
		str += "complex"
		return
	}

	seen[ty]++

	switch ty.Kind() {
	case reflect.Array:
		str += "array" + strconv.Itoa(ty.Len()) + looseGlob(ty.Elem(), seen)
		return

	case reflect.Chan:
		str += "chan" + looseGlob(ty.Elem(), seen)
		return

	case reflect.Func:
		for i := 0; i < ty.NumIn(); i++ {
			str += looseGlob(ty.In(i), seen)
		}

		// Without this, moving the first return value to the last input value
		// would look like an identical function.
		str += "out"

		for i := 0; i < ty.NumOut(); i++ {
			str += looseGlob(ty.Out(i), seen)
		}
		return

	case reflect.Interface:
		if ty.Name() != "" {
			str += ty.PkgPath() + "." + ty.Name()
			return
		}

		// Unlike struct's Field() method, Method() returns in lexographical order,
		// so we don't have to worry about order.
		for i := 0; i < ty.NumMethod(); i++ {
			m := ty.Method(i)
			str += m.Name + looseGlob(m.Type, seen)
		}

		return

	case reflect.Map:
		str += "map[" + looseGlob(ty.Key(), seen) + "]" + looseGlob(ty.Elem(), seen)
		return

	case reflect.Ptr:
		str += "*" + looseGlob(ty.Elem(), seen)
		return

	case reflect.Slice:
		str += "[]" + looseGlob(ty.Elem(), seen)
		return

	case reflect.String:
		str += "string"
		return

	case reflect.Struct:
		if ty.Name() != "" {
			str += ty.PkgPath() + "." + ty.Name()
			return
		}

		fields := structFields(ty)

		for _, field := range fields {
			str += field.Name + looseGlob(field.Type, seen)
		}

		return

	case reflect.UnsafePointer:
		str += "unsafepointer"
		return
	}

	// No kind was matched, a new kind must have been added to golang
	// and the library needs updating.
	panic(encio.NewError(
		encio.ErrBadType,
		fmt.Sprintf("%v is of an unknown kind.", ty),
		0,
	))
}

func strictGlob(ty reflect.Type, seen map[reflect.Type]int) (str string) {
	if seen == nil {
		seen = make(map[reflect.Type]int)
	}

	str += ty.Kind().String()
	str += ty.PkgPath() + ty.Name()

	switch ty.Kind() {
	case reflect.Invalid,
		reflect.Bool,
		reflect.Int,
		reflect.Int8,
		reflect.Int16,
		reflect.Int32,
		reflect.Int64,
		reflect.Uint,
		reflect.Uint8,
		reflect.Uint16,
		reflect.Uint32,
		reflect.Uint64,
		reflect.Uintptr,
		reflect.Float32,
		reflect.Float64,
		reflect.Complex64,
		reflect.Complex128:
		return
	}

	if seen[ty] > 1 {
		return "recursed"
	}
	seen[ty]++

	switch ty.Kind() {
	case reflect.Array:
		str += "array" + strconv.Itoa(ty.Len()) + strictGlob(ty.Elem(), seen)
		return

	case reflect.Chan:
		str += "chan" + strictGlob(ty.Elem(), seen)
		return

	case reflect.Func:
		for i := 0; i < ty.NumIn(); i++ {
			str += strictGlob(ty.In(i), seen)
		}

		// Without this, moving the first return value to the last input value
		// would look like an identical function.
		str += "out"

		for i := 0; i < ty.NumOut(); i++ {
			str += strictGlob(ty.Out(i), seen)
		}
		return

	case reflect.Interface:
		// Unlike struct's Field() method, Method() returns in lexographical order,
		// so we don't have to worry about order.
		for i := 0; i < ty.NumMethod(); i++ {
			m := ty.Method(i)
			str += m.Name + strictGlob(m.Type, seen)
		}

		return

	case reflect.Map:
		str += "map[" + strictGlob(ty.Key(), seen) + "]" + strictGlob(ty.Elem(), seen)
		return

	case reflect.Ptr:
		str += "*" + strictGlob(ty.Elem(), seen)
		return

	case reflect.Slice:
		str += "[]" + strictGlob(ty.Elem(), seen)
		return

	case reflect.String:
		str += "string"
		return

	case reflect.Struct:
		fields := structFields(ty)

		for _, field := range fields {
			str += field.Name + strictGlob(field.Type, seen)
		}

		return

	case reflect.UnsafePointer:
		str += "unsafepointer"
		return
	}

	// No kind was matched, a new kind must have been added to golang
	// and the library needs updating.
	panic(encio.NewError(
		encio.ErrBadType,
		fmt.Sprintf("%v is of an unknown kind.", ty),
		0,
	))
}
