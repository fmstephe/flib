package spscq

import (
	"fmt"
	"sync/atomic"
	"unsafe"

	"github.com/fmstephe/flib/fmath"
	"github.com/fmstephe/flib/fsync/fatomic"
	"github.com/fmstephe/flib/fsync/padded"
)

type PointerQ struct {
	paddedCounters
	ringBuffer  []unsafe.Pointer
	size        int64
	mask        int64
	_postbuffer padded.CacheBuffer
}

func NewPointerQ(size int64) *PointerQ {
	if !fmath.PowerOfTwo(size) {
		panic(fmt.Sprintf("Size must be a power of two, size = %d", size))
	}
	ringBuffer := padded.PointerSlice(int(size))
	q := &PointerQ{ringBuffer: ringBuffer, size: size, mask: size - 1}
	return q
}

func (q *PointerQ) WriteSingle(val unsafe.Pointer) bool {
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
	atomic.AddInt64(&q.write.Value, 1)
	return true
}

func (q *PointerQ) WriteBuffer(bufferSize int64) []unsafe.Pointer {
	write := q.write.Value
	idx := write & q.mask
	bufferSize = fmath.Min(bufferSize, q.size-idx)
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
	q.writeSize.Value = nxt - idx
	return q.ringBuffer[idx:nxt]
}

func (q *PointerQ) CommitWrite() {
	atomic.AddInt64(&q.write.Value, q.writeSize.Value)
	q.writeSize.Value = 0
}

func (q *PointerQ) CommitWriteLazy() {
	fatomic.LazyStore(&q.write.Value, q.write.Value+q.writeSize.Value)
	q.writeSize.Value = 0
}

func (q *PointerQ) ReadSingle() unsafe.Pointer {
	read := q.read.Value
	if read == q.writeCache.Value {
		q.writeCache.Value = atomic.LoadInt64(&q.write.Value)
		if read == q.writeCache.Value {
			q.readFail.Value++
			return nil
		}
	}
	val := q.ringBuffer[read&q.mask]
	atomic.AddInt64(&q.read.Value, 1)
	return val
}

func (q *PointerQ) ReadBuffer(bufferSize int64) []unsafe.Pointer {
	read := q.read.Value
	idx := read & q.mask
	bufferSize = fmath.Min(bufferSize, q.size-idx)
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
	q.readSize.Value = nxt - idx
	return q.ringBuffer[idx:nxt]
}

func (q *PointerQ) CommitRead() {
	atomic.AddInt64(&q.read.Value, q.readSize.Value)
	q.readSize.Value = 0
}

func (q *PointerQ) CommitReadLazy() {
	fatomic.LazyStore(&q.read.Value, q.read.Value+q.readSize.Value)
	q.readSize.Value = 0
}
