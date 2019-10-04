package gram

import (
	"sync/atomic"
	"unsafe"
)

// MinBufferSize is the minimum size for new buffers.
var MinBufferSize = 32

const (
	bpLocked = 1 << iota
	bpOff    = 1 << 16
	bpLen    = 1024
)

var bpState uint32
var buffers [bpLen][]byte

func getBuffer() []byte {
	s := bpState
	if s&bpLocked > 0 || s>>16 == 0 {
		return make([]byte, 0, MinBufferSize)
	}
	if atomic.CompareAndSwapUint32(&bpState, s, s&bpLocked) {
		b := buffers[(s>>16)-1]
		atomic.StoreUint32(&bpState, s-bpOff)
		if cap(b) < MinBufferSize {
			return make([]byte, 0, MinBufferSize)
		}
		return b[:0]
	}
	return make([]byte, 0, MinBufferSize)
}

func putBuffer(buff []byte) {
	for i := range buff {
		buff[i] = 0
	}
	s := bpState
	if s&bpLocked > 0 || s>>16 == bpLen-1 {
		return
	}
	if atomic.CompareAndSwapUint32(&bpState, s, s&bpLocked) {
		buffers[s>>16] = buff
		atomic.StoreUint32(&bpState, s+bpOff)
		return
	}
	return
}

// Sliced returns the offset and true if reslice is a re-slice of b.
func Sliced(buff, reslice []byte) (off int, sliced bool) {
	index := *(*uintptr)(unsafe.Pointer(&reslice))
	begin := *(*uintptr)(unsafe.Pointer(&buff))

	off = int(index - begin)
	c := cap(buff)
	if off > c {
		return c, false
	}
	return int(off), true
}

// SetCap returns the slice with the capacity set to n.
// Length is shrunk if neccecary.
func SetCap(buff []byte, n int) []byte {
	if len(buff) > n {
		buff = buff[:n]
	}
	cptr := (uintptr)(unsafe.Pointer(&buff)) + (wordSize * 2)
	*(*int)(unsafe.Pointer(cptr)) = n
	return buff
}
