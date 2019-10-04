package gram

// WriteSizeHeader is a helper function for writing a simple header to data containing its size.
// It reserves 4 bytes for a size header, and when finish is called,
// the number of bytes written to g after SizeHeader was called is written to the reserved space.
func WriteSizeHeader(g *Gram) (finish func()) {
	header := g.WriteLater(4)
	size := g.Size()
	return func() {
		header.WriteUint32(uint32(g.Size() - size))
	}
}

// ReadSizeHeader reads the 4 byte header written by WriteSizeHeader,
// returning a gram for reading the written data.
func ReadSizeHeader(g *Gram) *Gram {
	l := g.ReadUint32()
	return g.LimitReader(int(l))
}
