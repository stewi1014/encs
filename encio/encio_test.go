package encio_test

import (
	"bytes"
	"fmt"
	"io"
	"math/rand"

	"github.com/stewi1014/encs/encio"
)

func randomBytes(rng *rand.Rand, maxLen int) []byte {
	buff := make([]byte, 8+rng.Intn(maxLen))
	rng.Read(buff)
	return buff
}

func testWrite(w io.Writer, seed, payloads, maxLen int) error {
	rng := rand.New(rand.NewSource(int64(seed)))

	for i := 0; i < payloads; i++ {
		buff := randomBytes(rng, maxLen)

		n, err := w.Write(buff)
		if n != len(buff) || err != nil {
			return fmt.Errorf("wrote %v of %v bytes with err %v", n, len(buff), err)
		}
	}

	return nil
}

func testRead(r io.Reader, seed, payloads, maxLen int) error {
	rng := rand.New(rand.NewSource(int64(seed)))

	for i := 0; i < payloads; i++ {
		buff := randomBytes(rng, maxLen)

		var rbuff encio.Buffer
		n, err := rbuff.ReadNFrom(r, len(buff))
		if n != len(buff) || err != nil {
			return fmt.Errorf("read %v of %v bytes with err %v", n, len(buff), err)
		}

		if !bytes.Equal(buff, rbuff) {
			return fmt.Errorf("read data is different wrom written data")
		}
	}

	return nil
}
