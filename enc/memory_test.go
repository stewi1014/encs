package enc_test

import (
	"bytes"
	"testing"
	"unsafe"

	"github.com/stewi1014/encs/enc"
)

func TestMemory(t *testing.T) {
	var a1, a2 [64]byte
	buff := new(bytes.Buffer)
	m := enc.NewMemory(64)

	for i := range a1 {
		a1[i] = byte(i)
	}

	err := m.Encode(unsafe.Pointer(&a1), buff)
	if err != nil {
		t.Errorf("encode error; %v", err)
	}

	err = m.Decode(unsafe.Pointer(&a2), buff)
	if err != nil {
		t.Errorf("decode error; %v", err)
	}

	for i := range a1 {
		if a1[i] != a2[i] {
			t.Fatalf("wrong byte; have %v, want %v", a2[i], a1[i])
		}
	}

	err = m.Decode(nil, buff)
	if err == nil {
		t.Fatalf("no error thrown with nil pointer")
	}
}
