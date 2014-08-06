package spscq

import (
	"fmt"
	"github.com/fmstephe/flib/fmath"
	"github.com/fmstephe/flib/fsync/fatomic"
	"github.com/fmstephe/flib/fsync/padded"
	"sync/atomic"
)

type commonQ struct {
	// Readonly Fields
	size       int64
	mask       int64
	_ropadding padded.CacheBuffer
	// Writer fields
	write     padded.Int64
	writeSize padded.Int64
	writeFail padded.Int64
	readCache padded.Int64
	// Reader fields
	read       padded.Int64
	readSize   padded.Int64
	readFail   padded.Int64
	writeCache padded.Int64
}

func newCommonQ(size int64) commonQ {
	if !fmath.PowerOfTwo(size) {
		panic(fmt.Sprintf("Size (%d) must be a power of two", size))
	}
	cq := commonQ{size: size, mask: size - 1}
	return cq
}

func (q *commonQ) writeBuffer(bufferSize int64) (int64, int64) {
	write := q.write.Value
	from := write & q.mask
	bufferSize = fmath.Min(bufferSize, q.size-from)
	writeTo := write + bufferSize
	readLimit := writeTo - q.size
	to := from + bufferSize
	if readLimit > q.readCache.Value {
		q.readCache.Value = atomic.LoadInt64(&q.read.Value)
		if readLimit > q.readCache.Value {
			to = q.readCache.Value & q.mask
		}
	}
	if from == to {
		q.writeFail.Value++
	}
	q.writeSize.Value = to - from
	return from, to
}

func (q *commonQ) CommitWrite() {
	atomic.AddInt64(&q.write.Value, q.writeSize.Value)
	q.writeSize.Value = 0
}

func (q *commonQ) CommitWriteLazy() {
	fatomic.LazyStore(&q.write.Value, q.write.Value+q.writeSize.Value)
	q.writeSize.Value = 0
}

func (q *commonQ) readBuffer(bufferSize int64) (int64, int64) {
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
	return idx, nxt
}

func (q *commonQ) CommitRead() {
	atomic.AddInt64(&q.read.Value, q.readSize.Value)
	q.readSize.Value = 0
}

func (q *commonQ) CommitReadLazy() {
	fatomic.LazyStore(&q.read.Value, q.read.Value+q.readSize.Value)
	q.readSize.Value = 0
}

func (c *commonQ) WriteFails() int64 {
	return atomic.LoadInt64(&c.writeFail.Value)
}

func (c *commonQ) ReadFails() int64 {
	return atomic.LoadInt64(&c.readFail.Value)
}
