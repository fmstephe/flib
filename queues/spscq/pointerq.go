package spscq

import (
	"fmt"
	"github.com/fmstephe/flib/fsync/padded"
	"sync/atomic"
	"unsafe"
)

type PointerQ struct {
	_1         padded.Int64
	read       int64
	writeCache int64
	_2         padded.Int64
	write      int64
	readCache  int64
	_3         padded.Int64
	// Read only
	ringBuffer []unsafe.Pointer
	size       int64
	mask       int64
	_4         padded.Int64
}

func NewPointerQ(size int64) *PointerQ {
	if !powerOfTwo(size) {
		panic(fmt.Sprintf("Size must be a power of two, size = %d", size))
	}
	ringBuffer := padded.PointerSlice(int(size))
	q := &PointerQ{ringBuffer: ringBuffer, size: size, mask: size - 1}
	return q
}

func (q *PointerQ) WriteSingle(val unsafe.Pointer) bool {
	write := q.write
	readLimit := write - q.size
	if readLimit == q.readCache {
		q.readCache = atomic.LoadInt64(&q.read)
		if readLimit == q.readCache {
			return false
		}
	}
	q.ringBuffer[write&q.mask] = val
	atomic.AddInt64(&q.write, 1)
	return true
}

func (q *PointerQ) WriteBuffer(bufferSize int64) []unsafe.Pointer {
	write := q.write
	idx := write & q.mask
	bufferSize = min(bufferSize, q.size-idx)
	writeTo := write + bufferSize
	readLimit := writeTo - q.size
	nxt := idx + bufferSize
	if readLimit > q.readCache {
		q.readCache = atomic.LoadInt64(&q.read)
		if readLimit > q.readCache {
			nxt = q.readCache & q.mask
		}
	}
	return q.ringBuffer[idx:nxt]
}

func (q *PointerQ) CommitWriteBuffer(writeSize int64) {
	atomic.AddInt64(&q.write, writeSize)
}

func (q *PointerQ) ReadSingle() unsafe.Pointer {
	read := q.read
	if read == q.writeCache {
		q.writeCache = atomic.LoadInt64(&q.write)
		if read == q.writeCache {
			return nil
		}
	}
	val := q.ringBuffer[read&q.mask]
	atomic.AddInt64(&q.read, 1)
	return val
}

func (q *PointerQ) ReadBuffer(bufferSize int64) []unsafe.Pointer {
	read := q.read
	idx := read & q.mask
	bufferSize = min(bufferSize, q.size-idx)
	readTo := read + bufferSize
	nxt := idx + bufferSize
	if readTo > q.writeCache {
		q.writeCache = atomic.LoadInt64(&q.write)
		if readTo > q.writeCache {
			nxt = q.writeCache & q.mask
		}
	}
	return q.ringBuffer[idx:nxt]
}

func (q *PointerQ) CommitReadBuffer(readSize int64) {
	atomic.AddInt64(&q.read, readSize)
}
