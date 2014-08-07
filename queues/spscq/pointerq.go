package spscq

import (
	"sync/atomic"
	"unsafe"

	"github.com/fmstephe/flib/fsync/fatomic"
	"github.com/fmstephe/flib/fsync/padded"
)

type PointerQ struct {
	_prebuffer padded.CacheBuffer
	commonQ
	_midbuffer  padded.CacheBuffer
	ringBuffer  []unsafe.Pointer
	_postbuffer padded.CacheBuffer
}

func NewPointerQ(size int64) *PointerQ {
	ringBuffer := padded.PointerSlice(int(size))
	q := &PointerQ{ringBuffer: ringBuffer, commonQ: newCommonQ(size)}
	return q
}

func (q *PointerQ) WriteSingle(val unsafe.Pointer) bool {
	b := q.writeSingle(val)
	if b {
		atomic.AddInt64(&q.write.Value, 1)
	}
	return b
}

func (q *PointerQ) WriteSingleLazy(val unsafe.Pointer) bool {
	b := q.writeSingle(val)
	if b {
		fatomic.LazyStore(&q.write.Value, q.write.Value+1)
	}
	return b
}

func (q *PointerQ) writeSingle(val unsafe.Pointer) bool {
	write := q.write.Value
	readLimit := write - q.size
	if readLimit == q.readCache.Value {
		q.readCache.Value = atomic.LoadInt64(&q.read.Value)
		if readLimit == q.readCache.Value {
			q.failedWrites.Value++
			return false
		}
	}
	q.ringBuffer[write&q.mask] = val
	return true
}

func (q *PointerQ) WriteBuffer(bufferSize int64) []unsafe.Pointer {
	from, to := q.writeBuffer(bufferSize)
	return q.ringBuffer[from:to]
}

func (q *PointerQ) ReadSingle() unsafe.Pointer {
	val := q.readSingle()
	if val != nil {
		atomic.AddInt64(&q.read.Value, 1)
	}
	return val
}

func (q *PointerQ) ReadSingleLazy() unsafe.Pointer {
	val := q.readSingle()
	if val != nil {
		fatomic.LazyStore(&q.read.Value, q.read.Value+1)
	}
	return val
}

func (q *PointerQ) readSingle() unsafe.Pointer {
	read := q.read.Value
	if read == q.writeCache.Value {
		q.writeCache.Value = atomic.LoadInt64(&q.write.Value)
		if read == q.writeCache.Value {
			q.failedReads.Value++
			return nil
		}
	}
	val := q.ringBuffer[read&q.mask]
	return val
}

func (q *PointerQ) ReadBuffer(bufferSize int64) []unsafe.Pointer {
	from, to := q.readBuffer(bufferSize)
	return q.ringBuffer[from:to]
}
