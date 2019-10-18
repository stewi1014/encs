package gram

import (
	"fmt"
	"math/bits"
	"sync"
	"sync/atomic"
)

func init() {
	// set pool new functions
	for i := 1; i < 32; i++ {
		index := i
		buffers[i].New = func() interface{} {
			return make([]byte, 1<<(index-1))
		}
	}
}

// GetBuffer returns a buffer with a cap of at least n and len of 0 from the pool.
// If it is faster to allocate instead of using the pool, it allocates instead.
func GetBuffer(n int) []byte {
	//check(uint64(n))

	i := uint(bits.Len(uint(n)))
	if n != 1<<i {
		i++
	}
	return buffers[i].Get().([]byte)[:0]
}

// PutBuffer places a buffer in the buffer pool.
// If the buffer is too small to make pooling efficient, it discards the buffer.
func PutBuffer(buff []byte) {
	i := bits.Len(uint(cap(buff)))
	buffers[i].Put(buff)
}

var (
	buffers [32]sync.Pool
)

// BufferPool provides a pool of buffers of the same size.
type BufferPool struct {
	mutex   sync.RWMutex
	state   uint64 //32bits offset, 32bits len
	buffers [][]byte

	// Size sets the minimum size for buffers returned by Get
	Size int
}

const (
	bpLen    = 1
	bpLenSub = 1<<64 - 1

	bpOff    = 1 << 32
	bpOffSub = 1<<64 - bpOff

	bpLenMask = 1<<32 - 1
)

// Get returns a buffer with a length of BufferPool.Size.
// If it is faster to allocate instead of pooling, it does so.
func (bp *BufferPool) Get() []byte {
	bp.mutex.RLock()
	state := atomic.AddUint64(&bp.state, bpOff)
	if state>>32 > state&bpLenMask {
		atomic.AddUint64(&bp.state, bpOffSub)
		bp.mutex.RUnlock()
		return make([]byte, 0, bp.Size)
	}
	buff := bp.buffers[(state>>32)-1]
	bp.mutex.RUnlock()
	return buff
}

// Put returns a buffer to the pool. If the buffer is too small to make pooling efficient,
// it discards the buffer.
func (bp *BufferPool) Put(buff []byte) {
	if cap(buff) != bp.Size {
		panic(fmt.Errorf("BufferPool given buffer of size %v, but want %v", cap(buff), bp.Size))
	}
	bp.mutex.RLock()
	state := atomic.AddUint64(&bp.state, bpLen)
	if state&bpLenMask >= uint64(len(bp.buffers)) {
		atomic.AddUint64(&bp.state, bpLenSub)
		bp.mutex.RUnlock()
		bp.putSlow(buff)
		return
	}

	bp.buffers[(state&bpLenMask)-1] = buff[:0]
	bp.mutex.RUnlock()
	return
}

func (bp *BufferPool) putSlow(buff []byte) {
	bp.mutex.Lock()

	l := uint(bp.state & bpLenMask)
	off := uint(bp.state >> 32)
	ur := l - off
	c := uint(cap(bp.buffers))

	// we only need ur+1 < c, but we let c get 4 times as large to avoid constantly sliding.
	if (ur+1)*4 < c { // slide down
		copy(bp.buffers, bp.buffers[off:])

	} else { // allocate
		if c == 0 {
			c = 64
		}
		nb := make([][]byte, c*2)
		copy(nb, bp.buffers[off:])
		bp.buffers = nb
	}

	bp.buffers[ur] = buff[:0]
	bp.state = uint64(ur + 1)
	bp.mutex.Unlock()
	return
}
