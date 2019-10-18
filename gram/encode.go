package gram

import "io"

// Writer is a method for writing grams.
// The message should be written to the gram, and then sent by calling send.
type Writer interface {
	Write() (g *Gram, send func() error)
}

// NewStreamWriter returns a new StreamWriter writing to w.
func NewStreamWriter(w io.Writer) StreamWriter {
	return StreamWriter{
		w: w,
	}
}

// StreamWriter is the most basic system for sending Grams on the wire.
// It writes a 32bit number for the size of the gram, and
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

// NewPromiscuousWriter creates a new PromiscuousWriter
func NewPromiscuousWriter(w io.Writer) PromiscuousWriter {
	return PromiscuousWriter{
		w: w,
	}
}

// PromiscuousWriter is a Writer that uses an escape byte to define the begining and end of payloads.
type PromiscuousWriter struct {
	w io.Writer
}

const (
	escapeByte = 200

	gramBegin = iota
	escapeEscape
)

func (pw PromiscuousWriter) Write() (*Gram, func() error) {
	g, write := WriteGram(pw.w)
	header := g.WriteLater(6)
	return g, func() error {
		header.WriteByte(escapeByte)
		header.WriteByte(gramBegin)
		header.WriteUint32(uint32(g.Size() - 6))

		i := new(Iter)
		i.Delta(2) // skip first 2 header bytes
		for i.Iter(g) {
			if g.Byte(i) == escapeByte {
				g.InsertAfter(i, 1)[0] = escapeEscape // escape the escape.
			}
		}

		return write()
	}
}
