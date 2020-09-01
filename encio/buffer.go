package encio

// Buffer is a byte slice with helpful methods.
type Buffer []byte

// Write implements io.Writer.
// It appends buff to the buffer.
func (b *Buffer) Write(buff []byte) (int, error) {
	copy((*b)[b.Grow(len(buff)):], buff)
	return len(buff), nil
}

// Grow grows the buffer by n bytes, returning the previous length of the buffer.
func (b *Buffer) Grow(n int) int {
	buff := *b
	l := len(buff)
	c := cap(buff)
	if l+n <= c {
		*b = buff[:l+n]
		return l
	}

	*b = make([]byte, l+n, c*2+n)
	copy(*b, buff)
	return l
}
