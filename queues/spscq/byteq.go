package spscq

import (
	"sync/atomic"

	"github.com/fmstephe/flib/fmath"
	"github.com/fmstephe/flib/fsync/fatomic"
	"github.com/fmstephe/flib/fsync/padded"
)

type ByteQ struct {
	_prebuffer padded.CacheBuffer
	commonQ
	_midbuffer  padded.CacheBuffer
	ringBuffer  []byte
	_postbuffer padded.CacheBuffer
}

func NewByteQ(size int64) *ByteQ {
	ringBuffer := padded.ByteSlice(int(size))
	q := &ByteQ{ringBuffer: ringBuffer, commonQ: newCommonQ(size)}
	return q
}

func (q *ByteQ) WriteBuffer(bufferSize int64) []byte {
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
		q.failedWrites.Value++
	}
	q.writeSize.Value = nxt - idx
	return q.ringBuffer[idx:nxt]
}

func (q *ByteQ) CommitWrite() {
	atomic.AddInt64(&q.write.Value, q.writeSize.Value)
	q.writeSize.Value = 0
}

func (q *ByteQ) CommitWriteLazy() {
	fatomic.LazyStore(&q.write.Value, q.write.Value+q.writeSize.Value)
	q.writeSize.Value = 0
}

func (q *ByteQ) ReadBuffer(bufferSize int64) []byte {
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
		q.failedReads.Value++
	}
	q.readSize.Value = nxt - idx
	return q.ringBuffer[idx:nxt]
}

func (q *ByteQ) CommitRead() {
	atomic.AddInt64(&q.read.Value, q.readSize.Value)
	q.readSize.Value = 0
}

func (q *ByteQ) CommitReadLazy() {
	fatomic.LazyStore(&q.read.Value, q.read.Value+q.readSize.Value)
	q.readSize.Value = 0
}
