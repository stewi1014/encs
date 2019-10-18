package enc_test

import (
	"bytes"
	"testing"

	"github.com/stewi1014/encs/enc"
)

// helper functions for testing

func checkSize(buff *bytes.Buffer, e enc.Encodable, t *testing.T) {
	s := e.Size()
	if s < 0 {
		return
	}

	if buff.Len() > s {
		t.Fatalf("reported size smaller than written bytes; reported %v but wrote %v bytes", s, buff.Len())
	}
}
