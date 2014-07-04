package spscq

import (
	"fmt"
	"github.com/fmstephe/flib/fsync/fatomic"
	"github.com/fmstephe/flib/fsync/padded"
	"sync/atomic"
	"unsafe"
)

type UnsafePointerQ struct {
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

func NewUnsafePointerQ(size int64) *UnsafePointerQ {
	if !powerOfTwo(size) {
		panic(fmt.Sprintf("Size must be a power of two, size = %d", size))
	}
	ringBuffer := padded.PointerSlice(int(size))
	q := &UnsafePointerQ{ringBuffer: ringBuffer, size: size, mask: size - 1}
	return q
}

func (q *UnsafePointerQ) WriteSingle(val unsafe.Pointer) bool {
	write := q.write
	readLimit := write - q.size
	if readLimit == q.readCache {
		q.readCache = atomic.LoadInt64(&q.read)
		if readLimit == q.readCache {
			return false
		}
	}
	q.ringBuffer[write&q.mask] = val
	fatomic.LazyStore(&q.write, q.write+1)
	return true
}

func (q *UnsafePointerQ) WriteBuffer(bufferSize int64) []unsafe.Pointer {
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

func (q *UnsafePointerQ) CommitWriteBuffer(writeSize int64) {
	fatomic.LazyStore(&q.write, q.write+writeSize)
}

func (q *UnsafePointerQ) ReadSingle() unsafe.Pointer {
	read := q.read
	if read == q.writeCache {
		q.writeCache = atomic.LoadInt64(&q.write)
		if read == q.writeCache {
			return nil
		}
	}
	val := q.ringBuffer[read&q.mask]
	fatomic.LazyStore(&q.read, q.read+1)
	return val
}

func (q *UnsafePointerQ) ReadBuffer(bufferSize int64) []unsafe.Pointer {
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

func (q *UnsafePointerQ) CommitReadBuffer(readSize int64) {
	fatomic.LazyStore(&q.read, q.read+readSize)
}
