package encode_test

import (
	"fmt"
	"io/ioutil"
	"reflect"
	"testing"
	"unsafe"

	"github.com/maxatome/go-testdeep/td"
	"github.com/stewi1014/encs/encodable"
	"github.com/stewi1014/encs/encode"
)

type varintTestCase struct {
	desc string
	enc  interface{}
	dec  interface{}
}

func testVarintExplodeint(n int) []varintTestCase {
	encs := []interface{}{
		int(n),
		int8(n),
		int16(n),
		int32(n),
		int64(n),
		uint(n),
		uint8(n),
		uint16(n),
		uint32(n),
		uint64(n),
		uintptr(n),
		float32(n),
		float64(n),
	}

	tcs := make([]varintTestCase, len(encs))

	for j := range encs {
		tcs[j] = varintTestCase{
			desc: fmt.Sprintf("%T (%v) to %T (%v)", n, n, encs[j], encs[j]),
			enc:  n,
			dec:  encs[j],
		}
	}

	return tcs
}

func testVarintExplodeint8(n int8) []varintTestCase {
	encs := []interface{}{
		int(n),
		int8(n),
		int16(n),
		int32(n),
		int64(n),
		uint(n),
		uint8(n),
		uint16(n),
		uint32(n),
		uint64(n),
		uintptr(n),
		float32(n),
		float64(n),
	}

	tcs := make([]varintTestCase, len(encs))

	for j := range encs {
		tcs[j] = varintTestCase{
			desc: fmt.Sprintf("%T (%v) to %T (%v)", n, n, encs[j], encs[j]),
			enc:  n,
			dec:  encs[j],
		}
	}

	return tcs
}

func testVarintExplodeint16(n int16) []varintTestCase {
	encs := []interface{}{
		int(n),
		int8(n),
		int16(n),
		int32(n),
		int64(n),
		uint(n),
		uint8(n),
		uint16(n),
		uint32(n),
		uint64(n),
		uintptr(n),
		float32(n),
		float64(n),
	}

	tcs := make([]varintTestCase, len(encs))

	for j := range encs {
		tcs[j] = varintTestCase{
			desc: fmt.Sprintf("%T (%v) to %T (%v)", n, n, encs[j], encs[j]),
			enc:  n,
			dec:  encs[j],
		}
	}

	return tcs
}

func testVarintExplodeint32(n int32) []varintTestCase {
	encs := []interface{}{
		int(n),
		int8(n),
		int16(n),
		int32(n),
		int64(n),
		uint(n),
		uint8(n),
		uint16(n),
		uint32(n),
		uint64(n),
		uintptr(n),
		float32(n),
		float64(n),
	}

	tcs := make([]varintTestCase, len(encs))

	for j := range encs {
		tcs[j] = varintTestCase{
			desc: fmt.Sprintf("%T (%v) to %T (%v)", n, n, encs[j], encs[j]),
			enc:  n,
			dec:  encs[j],
		}
	}

	return tcs
}

func testVarintExplodeint64(n int64) []varintTestCase {
	encs := []interface{}{
		int(n),
		int8(n),
		int16(n),
		int32(n),
		int64(n),
		uint(n),
		uint8(n),
		uint16(n),
		uint32(n),
		uint64(n),
		uintptr(n),
		float32(n),
		float64(n),
	}

	tcs := make([]varintTestCase, len(encs))

	for j := range encs {
		tcs[j] = varintTestCase{
			desc: fmt.Sprintf("%T (%v) to %T (%v)", n, n, encs[j], encs[j]),
			enc:  n,
			dec:  encs[j],
		}
	}

	return tcs
}

func testVarintExplodeuint(n uint) []varintTestCase {
	encs := []interface{}{
		int(n),
		int8(n),
		int16(n),
		int32(n),
		int64(n),
		uint(n),
		uint8(n),
		uint16(n),
		uint32(n),
		uint64(n),
		uintptr(n),
		float32(n),
		float64(n),
	}

	tcs := make([]varintTestCase, len(encs))

	for j := range encs {
		tcs[j] = varintTestCase{
			desc: fmt.Sprintf("%T (%v) to %T (%v)", n, n, encs[j], encs[j]),
			enc:  n,
			dec:  encs[j],
		}
	}

	return tcs
}

func testVarintExplodeuint8(n uint8) []varintTestCase {
	encs := []interface{}{
		int(n),
		int8(n),
		int16(n),
		int32(n),
		int64(n),
		uint(n),
		uint8(n),
		uint16(n),
		uint32(n),
		uint64(n),
		uintptr(n),
		float32(n),
		float64(n),
	}

	tcs := make([]varintTestCase, len(encs))

	for j := range encs {
		tcs[j] = varintTestCase{
			desc: fmt.Sprintf("%T (%v) to %T (%v)", n, n, encs[j], encs[j]),
			enc:  n,
			dec:  encs[j],
		}
	}

	return tcs
}
func testVarintExplodeuint16(n uint16) []varintTestCase {
	encs := []interface{}{
		int(n),
		int8(n),
		int16(n),
		int32(n),
		int64(n),
		uint(n),
		uint8(n),
		uint16(n),
		uint32(n),
		uint64(n),
		uintptr(n),
		float32(n),
		float64(n),
	}

	tcs := make([]varintTestCase, len(encs))

	for j := range encs {
		tcs[j] = varintTestCase{
			desc: fmt.Sprintf("%T (%v) to %T (%v)", n, n, encs[j], encs[j]),
			enc:  n,
			dec:  encs[j],
		}
	}

	return tcs
}

func testVarintExplodeuint32(n uint32) []varintTestCase {
	encs := []interface{}{
		int(n),
		int8(n),
		int16(n),
		int32(n),
		int64(n),
		uint(n),
		uint8(n),
		uint16(n),
		uint32(n),
		uint64(n),
		uintptr(n),
		float32(n),
		float64(n),
	}

	tcs := make([]varintTestCase, len(encs))

	for j := range encs {
		tcs[j] = varintTestCase{
			desc: fmt.Sprintf("%T (%v) to %T (%v)", n, n, encs[j], encs[j]),
			enc:  n,
			dec:  encs[j],
		}
	}

	return tcs
}

func testVarintExplodeuint64(n uint64) []varintTestCase {
	encs := []interface{}{
		int(n),
		int8(n),
		int16(n),
		int32(n),
		int64(n),
		uint(n),
		uint8(n),
		uint16(n),
		uint32(n),
		uint64(n),
		uintptr(n),
		float32(n),
		float64(n),
	}

	tcs := make([]varintTestCase, len(encs))

	for j := range encs {
		tcs[j] = varintTestCase{
			desc: fmt.Sprintf("%T (%v) to %T (%v)", n, n, encs[j], encs[j]),
			enc:  n,
			dec:  encs[j],
		}
	}

	return tcs
}

func testVarintExplodefloat32(n float32) []varintTestCase {
	encs := []interface{}{
		int(n),
		int8(n),
		int16(n),
		int32(n),
		int64(n),
		uint(n),
		uint8(n),
		uint16(n),
		uint32(n),
		uint64(n),
		uintptr(n),
		float32(n),
		float64(n),
	}

	tcs := make([]varintTestCase, len(encs))

	for j := range encs {
		tcs[j] = varintTestCase{
			desc: fmt.Sprintf("%T (%v) to %T (%v)", n, n, encs[j], encs[j]),
			enc:  n,
			dec:  encs[j],
		}
	}

	return tcs
}

func testVarintExplodefloat64(n float64) []varintTestCase {
	encs := []interface{}{
		int(n),
		int8(n),
		int16(n),
		int32(n),
		int64(n),
		uint(n),
		uint8(n),
		uint16(n),
		uint32(n),
		uint64(n),
		uintptr(n),
		float32(n),
		float64(n),
	}

	tcs := make([]varintTestCase, len(encs))

	for j := range encs {
		tcs[j] = varintTestCase{
			desc: fmt.Sprintf("%T (%v) to %T (%v)", n, n, encs[j], encs[j]),
			enc:  n,
			dec:  encs[j],
		}
	}

	return tcs
}

func testVarintExplodeuintptr(n uintptr) []varintTestCase {
	encs := []interface{}{
		int(n),
		int8(n),
		int16(n),
		int32(n),
		int64(n),
		uint(n),
		uint8(n),
		uint16(n),
		uint32(n),
		uint64(n),
		uintptr(n),
		float32(n),
		float64(n),
	}

	tcs := make([]varintTestCase, len(encs))

	for j := range encs {
		tcs[j] = varintTestCase{
			desc: fmt.Sprintf("%T (%v) to %T (%v)", n, n, encs[j], encs[j]),
			enc:  n,
			dec:  encs[j],
		}
	}

	return tcs
}

func testVarintMetaExplodeint64(n int64) []varintTestCase {
	var testCases []varintTestCase
	testCases = append(testCases, testVarintExplodeint(int(n))...)
	testCases = append(testCases, testVarintExplodeint8(int8(n))...)
	testCases = append(testCases, testVarintExplodeint16(int16(n))...)
	testCases = append(testCases, testVarintExplodeint32(int32(n))...)
	testCases = append(testCases, testVarintExplodeint64(int64(n))...)
	testCases = append(testCases, testVarintExplodeuint(uint(n))...)
	testCases = append(testCases, testVarintExplodeuint8(uint8(n))...)
	testCases = append(testCases, testVarintExplodeuint16(uint16(n))...)
	testCases = append(testCases, testVarintExplodeuint32(uint32(n))...)
	testCases = append(testCases, testVarintExplodeuint64(uint64(n))...)
	testCases = append(testCases, testVarintExplodeuintptr(uintptr(n))...)
	testCases = append(testCases, testVarintExplodefloat32(float32(n))...)
	testCases = append(testCases, testVarintExplodefloat64(float64(n))...)
	return testCases
}

func testVarintMetaExplodeuint64(n uint64) []varintTestCase {
	var testCases []varintTestCase
	testCases = append(testCases, testVarintExplodeint(int(n))...)
	testCases = append(testCases, testVarintExplodeint8(int8(n))...)
	testCases = append(testCases, testVarintExplodeint16(int16(n))...)
	testCases = append(testCases, testVarintExplodeint32(int32(n))...)
	testCases = append(testCases, testVarintExplodeint64(int64(n))...)
	testCases = append(testCases, testVarintExplodeuint(uint(n))...)
	testCases = append(testCases, testVarintExplodeuint8(uint8(n))...)
	testCases = append(testCases, testVarintExplodeuint16(uint16(n))...)
	testCases = append(testCases, testVarintExplodeuint32(uint32(n))...)
	testCases = append(testCases, testVarintExplodeuint64(uint64(n))...)
	testCases = append(testCases, testVarintExplodeuintptr(uintptr(n))...)
	testCases = append(testCases, testVarintExplodefloat32(float32(n))...)
	testCases = append(testCases, testVarintExplodefloat64(float64(n))...)
	return testCases
}

func testVarintMetaExplodefloat64(n float64) []varintTestCase {
	var testCases []varintTestCase
	testCases = append(testCases, testVarintExplodeint(int(n))...)
	testCases = append(testCases, testVarintExplodeint8(int8(n))...)
	testCases = append(testCases, testVarintExplodeint16(int16(n))...)
	testCases = append(testCases, testVarintExplodeint32(int32(n))...)
	testCases = append(testCases, testVarintExplodeint64(int64(n))...)
	testCases = append(testCases, testVarintExplodeuint(uint(n))...)
	testCases = append(testCases, testVarintExplodeuint8(uint8(n))...)
	testCases = append(testCases, testVarintExplodeuint16(uint16(n))...)
	testCases = append(testCases, testVarintExplodeuint32(uint32(n))...)
	testCases = append(testCases, testVarintExplodeuint64(uint64(n))...)
	testCases = append(testCases, testVarintExplodeuintptr(uintptr(n))...)
	testCases = append(testCases, testVarintExplodefloat32(float32(n))...)
	testCases = append(testCases, testVarintExplodefloat64(float64(n))...)
	return testCases
}

func getVarintTestCases() (testCases []varintTestCase) {
	floatTests := []float64{
		-0.01,
		234506989860243564903872533333333,
		0, 1, float64(1<<64 - 1),
	}
	for _, f := range floatTests {
		testCases = append(testCases, testVarintMetaExplodefloat64(f)...)
	}

	intTests := []int64{
		1,
		(1 << 7) - 1, (1 << 7), (1 << 7) + 1,
		(1 << 15) - 1, (1 << 15), (1 << 15) + 1,
		(1 << 31) - 1, (1 << 31), (1 << 31) + 1,
		(1 << 63) - 1,
		0,
		-1,
		(-1 << 7) - 1, (-1 << 7), (-1 << 7) + 1,
		(-1 << 15) - 1, (-1 << 15), (-1 << 15) + 1,
		(-1 << 15) - 1, (-1 << 31), (-1 << 31) + 1,
		(-1 << 63), (-1 << 63) + 1,
	}

	for i := 0; i < 63; i++ {
		intTests = append(intTests, 1<<i, -1<<i)
	}

	for _, i := range intTests {
		testCases = append(testCases, testVarintMetaExplodeint64(i)...)
	}

	uintTests := []uint64{
		0, 1,
		254, 255, 256,
		(1 << 64) - 1, (1 << 63) + 1, (1 << 63),
	}
	for _, i := range uintTests {
		testCases = append(testCases, testVarintMetaExplodeuint64(i)...)
	}

	return
}

func TestVarint(t *testing.T) {
	testCases := getVarintTestCases()

	// Use a caching source so Varint is reused.
	src := encodable.NewCachingSource(encodable.SourceFromFunc(func(ty reflect.Type, source encodable.Source) encodable.Encodable {
		return encode.NewVarint(ty)
	}))

	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			decodeValue := reflect.New(reflect.TypeOf(tC.dec)).Elem()
			encodedValue := reflect.New(reflect.TypeOf(tC.enc)).Elem()
			encodedValue.Set(reflect.ValueOf(tC.enc))

			enc := src.NewEncodable(encodedValue.Type(), nil)
			dec := src.NewEncodable(decodeValue.Type(), nil)

			runTestNoErr(encodedValue, decodeValue, *enc, *dec, t)

			td.Cmp(t, decodeValue.Interface(), tC.dec)
		})
	}
}

var (
	int64Sink   int64
	uint16Sink  uint16
	float64Sink float64
)

func BenchmarkVarintEncode(b *testing.B) {
	b.Run("int64", func(b *testing.B) {
		send := int64(123456789)

		enc := encode.NewVarint(reflect.TypeOf(send))

		for i := 0; i < b.N; i++ {
			if err := enc.Encode(unsafe.Pointer(&send), ioutil.Discard); err != nil {
				b.Fatal(err)
			}
		}
	})
	b.Run("uint16", func(b *testing.B) {
		send := uint16(12349)

		enc := encode.NewVarint(reflect.TypeOf(send))

		for i := 0; i < b.N; i++ {
			if err := enc.Encode(unsafe.Pointer(&send), ioutil.Discard); err != nil {
				b.Fatal(err)
			}
		}
	})
	b.Run("float64", func(b *testing.B) {
		send := float64(123456789)

		enc := encode.NewVarint(reflect.TypeOf(send))

		for i := 0; i < b.N; i++ {
			if err := enc.Encode(unsafe.Pointer(&send), ioutil.Discard); err != nil {
				b.Fatal(err)
			}
		}
	})
}

func BenchmarkVarintDecode(b *testing.B) {
	int64Num := 123456789
	uint16Num := 23456
	float64Num := 77492.223

	int64Buffer := new(buffer)
	uint16Buffer := new(buffer)
	float64Buffer := new(buffer)

	int64Enc := encode.NewVarint(reflect.TypeOf(int64(0)))
	uint16Enc := encode.NewVarint(reflect.TypeOf(uint16(0)))
	float64Enc := encode.NewVarint(reflect.TypeOf(float64(0)))

	num := 10000

	for i := 0; i < num; i++ {
		if err := int64Enc.Encode(unsafe.Pointer(&int64Num), int64Buffer); err != nil {
			b.Fatal(err)
		}
		if err := uint16Enc.Encode(unsafe.Pointer(&uint16Num), uint16Buffer); err != nil {
			b.Fatal(err)
		}
		if err := float64Enc.Encode(unsafe.Pointer(&float64Num), float64Buffer); err != nil {
			b.Fatal(err)
		}
	}

	b.Run("int64 to int64", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			chunk := b.N
			if chunk > num {
				chunk = num
			}

			for j := 0; j < chunk; j++ {
				if err := int64Enc.Decode(unsafe.Pointer(&int64Sink), int64Buffer); err != nil {
					b.Fatal(err)
				}
			}

			int64Buffer.Reset()
			i += chunk
		}
	})

	b.Run("int64 to uint16", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			chunk := b.N
			if chunk > num {
				chunk = num
			}

			for j := 0; j < chunk; j++ {
				if err := uint16Enc.Decode(unsafe.Pointer(&uint16Sink), int64Buffer); err != nil {
					b.Fatal(err)
				}
			}

			int64Buffer.Reset()
			i += chunk
		}
	})

	b.Run("int64 to float64", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			chunk := b.N
			if chunk > num {
				chunk = num
			}

			for j := 0; j < chunk; j++ {
				if err := float64Enc.Decode(unsafe.Pointer(&float64Sink), int64Buffer); err != nil {
					b.Fatal(err)
				}
			}

			int64Buffer.Reset()
			i += chunk
		}
	})

	b.Run("uint16 to int64", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			chunk := b.N
			if chunk > num {
				chunk = num
			}

			for j := 0; j < chunk; j++ {
				if err := int64Enc.Decode(unsafe.Pointer(&int64Sink), uint16Buffer); err != nil {
					b.Fatal(err)
				}
			}

			uint16Buffer.Reset()
			i += chunk
		}
	})

	b.Run("uint16 to uint16", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			chunk := b.N
			if chunk > num {
				chunk = num
			}

			for j := 0; j < chunk; j++ {
				if err := uint16Enc.Decode(unsafe.Pointer(&uint16Sink), uint16Buffer); err != nil {
					b.Fatal(err)
				}
			}

			uint16Buffer.Reset()
			i += chunk
		}
	})

	b.Run("uint16 to float64", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			chunk := b.N
			if chunk > num {
				chunk = num
			}

			for j := 0; j < chunk; j++ {
				if err := float64Enc.Decode(unsafe.Pointer(&float64Sink), uint16Buffer); err != nil {
					b.Fatal(err)
				}
			}

			uint16Buffer.Reset()
			i += chunk
		}
	})

	b.Run("float64 to int64", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			chunk := b.N
			if chunk > num {
				chunk = num
			}

			for j := 0; j < chunk; j++ {
				if err := int64Enc.Decode(unsafe.Pointer(&int64Sink), float64Buffer); err != nil {
					b.Fatal(err)
				}
			}

			float64Buffer.Reset()
			i += chunk
		}
	})

	b.Run("float64 to uint16", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			chunk := b.N
			if chunk > num {
				chunk = num
			}

			for j := 0; j < chunk; j++ {
				if err := uint16Enc.Decode(unsafe.Pointer(&uint16Sink), float64Buffer); err != nil {
					b.Fatal(err)
				}
			}

			float64Buffer.Reset()
			i += chunk
		}
	})

	b.Run("float64 to float64", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			chunk := b.N
			if chunk > num {
				chunk = num
			}

			for j := 0; j < chunk; j++ {
				if err := float64Enc.Decode(unsafe.Pointer(&float64Sink), float64Buffer); err != nil {
					b.Fatal(err)
				}
			}

			float64Buffer.Reset()
			i += chunk
		}
	})
}
