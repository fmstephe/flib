package spscq

import (
	"fmt"
	"sync/atomic"
	"unsafe"

	"github.com/fmstephe/flib/fsync/fatomic"
	"github.com/fmstephe/flib/fsync/padded"
)

type UnsafePointerQ struct {
	_prebuffer  padded.CacheBuffer
	read        padded.Int64
	readFail    padded.Int64
	writeCache  padded.Int64
	write       padded.Int64
	writeFail   padded.Int64
	readCache   padded.Int64
	_midbuffer  padded.CacheBuffer
	ringBuffer  []unsafe.Pointer
	size        int64
	mask        int64
	_postbuffer padded.CacheBuffer
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
	write := q.write.Value
	readLimit := write - q.size
	if readLimit == q.readCache.Value {
		q.readCache.Value = atomic.LoadInt64(&q.read.Value)
		if readLimit == q.readCache.Value {
			q.writeFail.Value++
			return false
		}
	}
	q.ringBuffer[write&q.mask] = val
	fatomic.LazyStore(&q.write.Value, q.write.Value+1)
	return true
}

func (q *UnsafePointerQ) WriteBuffer(bufferSize int64) []unsafe.Pointer {
	write := q.write.Value
	idx := write & q.mask
	bufferSize = min(bufferSize, q.size-idx)
	writeTo := write + bufferSize
	readLimit := writeTo - q.size
	nxt := idx + bufferSize
	if readLimit > q.readCache.Value {
		q.readCache.Value = atomic.LoadInt64(&q.read.Value)
		if readLimit > q.readCache.Value {
			nxt = q.readCache.Value & q.mask
		}
	}
	if idx == nxt {
		q.writeFail.Value++
	}
	return q.ringBuffer[idx:nxt]
}

func (q *UnsafePointerQ) CommitWriteBuffer(writeSize int64) {
	fatomic.LazyStore(&q.write.Value, q.write.Value+writeSize)
}

func (q *UnsafePointerQ) ReadSingle() unsafe.Pointer {
	read := q.read.Value
	if read == q.writeCache.Value {
		q.writeCache.Value = atomic.LoadInt64(&q.write.Value)
		if read == q.writeCache.Value {
			q.readFail.Value++
			return nil
		}
	}
	val := q.ringBuffer[read&q.mask]
	fatomic.LazyStore(&q.read.Value, q.read.Value+1)
	return val
}

func (q *UnsafePointerQ) ReadBuffer(bufferSize int64) []unsafe.Pointer {
	read := q.read.Value
	idx := read & q.mask
	bufferSize = min(bufferSize, q.size-idx)
	readTo := read + bufferSize
	nxt := idx + bufferSize
	if readTo > q.writeCache.Value {
		q.writeCache.Value = atomic.LoadInt64(&q.write.Value)
		if readTo > q.writeCache.Value {
			nxt = q.writeCache.Value & q.mask
		}
	}
	if idx == nxt {
		q.readFail.Value++
	}
	return q.ringBuffer[idx:nxt]
}

func (q *UnsafePointerQ) CommitReadBuffer(readSize int64) {
	fatomic.LazyStore(&q.read.Value, q.read.Value+readSize)
}
