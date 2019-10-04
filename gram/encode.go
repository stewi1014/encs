package gram

import "io"

// Writer is a method for writing grams.
// The message should be written to the gram, and then sent by calling send.
type Writer interface {
	Write() (g *Gram, send func() error)
}

func NewStreamWriter(w io.Writer) StreamWriter {
	return StreamWriter{
		w: w,
	}
}

type StreamWriter struct {
	w io.Writer
}

func (sw StreamWriter) Write() (*Gram, func() error) {
	g, write := WriteGram(sw.w)
	header := g.WriteLater(4)
	return g, func() error {
		header.WriteUint32(uint32(g.Size() - 4))
		return write()
	}
}
