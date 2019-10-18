package gram

// Iter is an integer, used with gram for iteration functions.
// It is offset by 1 to allow "i := new(Iter); for g.Next(i) {...".
// 'new(Iter)' is an appropriate initialisation to iterate from 0.
type Iter int

// Delta shifts the iterator by n
func (i *Iter) Delta(n int) {
	*i += Iter(n)
}

// I returns the true index of the iterator.
func (i *Iter) I() int {
	if *i == 0 {
		return 0
	}
	return int(*i) - 1
}

// Iter iterates over bytes in a gram.
func (i *Iter) Iter(g *Gram) bool {
	if len(g.buff) > int(*i) {
		*i++
		return true
	}
	return false
}

// IterTo iterates over bytes in a gram until the index stop.
func (i *Iter) IterTo(g *Gram, stop int) bool {
	return i.Iter(g) && i.I() < stop
}

// Byte returns the byte at the iteration index.
func (g *Gram) Byte(i *Iter) byte {
	return g.buff[i.I()]
}

// InsertBefore inserts n bytes before the current byte,
// returning a slice for the inserted bytes.
// The iterator stays on the current byte.
func (g *Gram) InsertBefore(i *Iter, n int) []byte {
	g.slide(i.I(), n)
	buff := g.buff[i.I() : i.I()+n]
	i.Delta(n)
	return buff
}

// InsertAfter inserts n bytes after the current byte,
// returning a slice for the inserted bytes.
// The iterator shifts to the byte before the end of the inserted area (possibly nonexistent).
func (g *Gram) InsertAfter(i *Iter, n int) []byte {
	g.slide(i.I()+1, n)
	buff := g.buff[i.I()+1 : i.I()+1+n]
	i.Delta(n)
	return buff
}

// Remove removes the byte at the current iteration
func (g *Gram) Remove(i *Iter) {
	copy(g.buff[:i.I()], g.buff[i.I()+1:])
	g.buff = g.buff[:len(g.buff)-1]
}
