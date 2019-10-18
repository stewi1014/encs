package gram

import (
	"fmt"
	"testing"
)

var bufferSizeBenchmarks = []int{
	1, 10, 25, 50, 200, 1000, 2000000,
}

var buffSink []byte

func BenchmarkPool(b *testing.B) {
	for _, n := range bufferSizeBenchmarks {
		benchmarkPool(b, n)
	}
}

func benchmarkPool(b *testing.B, n int) {
	b.Run(fmt.Sprintf("BenchmarkPool-%v", n), func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			buffSink = GetBuffer(n)
			PutBuffer(buffSink)
		}
	})
}

func BenchmarkBufferPool(b *testing.B) {
	for _, n := range bufferSizeBenchmarks {
		benchmarkBufferPool(b, n)
	}
}

func benchmarkBufferPool(b *testing.B, n int) {
	b.Run(fmt.Sprintf("BenchmarkBufferPool-%v", n), func(b *testing.B) {
		bp := new(BufferPool)
		bp.Size = int(n)
		for i := 0; i < b.N; i++ {
			buffSink = bp.Get()
			bp.Put(buffSink)
		}
	})
}

func TestBufferPool(t *testing.T) {
	bp := new(BufferPool)
	bp.Size = 20
	for i := 0; i < 100; i++ {
		buffSink = bp.Get()
		if cap(buffSink) != 20 {
			t.Fatalf("buffer size wrong; got %v, wanted %v", cap(buffSink), bp.Size)
		}
		bp.Put(buffSink)
	}
}

func BenchmarkAllocate(b *testing.B) {
	for _, n := range bufferSizeBenchmarks {
		benchmarkAllocate(b, n)
	}
}

func benchmarkAllocate(b *testing.B, n int) {
	b.Run(fmt.Sprintf("BenchmarkAllocate-%v", n), func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			buffSink = make([]byte, n)
		}
	})
}
