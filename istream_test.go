package gneg

import (
	"bytes"
	"encoding/gob"
	"os"
	"reflect"
	"testing"
)

func TestMain(m *testing.M) {

	os.Exit(m.Run())
}

type testMarshaller struct {
	name string
}

func (e *testMarshaller) MarshalBinary() ([]byte, error) {
	return []byte(e.name), nil
}

func (e *testMarshaller) UnmarshalBinary(buff []byte) error {
	e.name = string(buff)
	return nil
}

type testStruct struct {
	Name   string
	Number int
}

var testValues = []interface{}{
	int(1),
	uint64(1<<64 - 1),
	float32(3.23),
	float64(4.112),
	[]int{1, 2, 3, 4, 5},
	[]float64{5, 5, 5, 5, 5, 5, 5, 5},
	"Hello World!",
	true,
	complex(234, 55),
	map[string]int{
		"apples":  3,
		"oranges": 25565,
	},
	&testMarshaller{name: "Hello again world!"},
	&testStruct{
		Name:   "Example Struct",
		Number: 15,
	},
}

func TestStream(t *testing.T) {
	Register(testMarshaller{})
	Register(testStruct{})
	buff := new(bytes.Buffer)

	enc := NewEncoder(buff, nil)
	dec := NewDecoder(buff, nil)

	var decoded interface{}
	for _, v := range testValues {
		err := enc.Encode(v)
		if err != nil {
			t.Error(err)
			return
		}

		err = dec.Decode(&decoded)
		if err != nil {
			t.Error(err)
			return
		}

		if !reflect.DeepEqual(v, decoded) {
			t.Errorf("encoding %v, decoding got %v", v, decoded)
		}
	}

}

func BenchmarkStream(b *testing.B) {
	buff := new(bytes.Buffer)
	enc := NewEncoder(buff, nil)
	dec := NewDecoder(buff, nil)

	var v interface{}
	j := 0
	for i := 0; i < b.N; i++ {
		enc.Encode(testValues[j])
		dec.Decode(&v)
		buff.Reset()

		j++
		if j == len(testValues) {
			j = 0
		}
	}
}

func BenchmarkGob(b *testing.B) {
	buff := new(bytes.Buffer)
	enc := gob.NewEncoder(buff)
	dec := gob.NewDecoder(buff)

	var v interface{}
	j := 0
	for i := 0; i < b.N; i++ {
		enc.Encode(testValues[j])
		dec.Decode(&v)
		buff.Reset()

		j++
		if j == len(testValues) {
			j = 0
		}
	}
}
