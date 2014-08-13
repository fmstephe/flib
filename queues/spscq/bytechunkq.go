package spscq

import (
	"fmt"
	"sync/atomic"

	"github.com/fmstephe/flib/fsync/fatomic"
	"github.com/fmstephe/flib/fsync/padded"
)

type ByteChunkQ struct {
	_prebuffer padded.CacheBuffer
	commonQ
	_midbuffer  padded.CacheBuffer
	ringBuffer  []byte
	chunk       int64
	_postbuffer padded.CacheBuffer
}

func NewByteChunkQ(size int64, chunk int64) *ByteChunkQ {
	if size%chunk != 0 {
		panic(fmt.Sprintf("Size must be neatly divisible by chunk, (size) %d rem (chunk) %d = %d", size, chunk, size%chunk))
	}
	ringBuffer := padded.ByteSlice(int(size))
	q := &ByteChunkQ{ringBuffer: ringBuffer, commonQ: newCommonQ(size), chunk: chunk}
	return q
}

func (q *ByteChunkQ) AcquireWrite() []byte {
	chunk := q.chunk
	write := q.write.Value
	writeTo := write + chunk
	readLimit := writeTo - q.size
	if readLimit > q.readCache.Value {
		q.readCache.Value = atomic.LoadInt64(&q.read.Value)
		if readLimit > q.readCache.Value {
			q.failedWrites.Value++
			return nil
		}
	}
	idx := write & q.mask
	nxt := idx + chunk
	return q.ringBuffer[idx:nxt]
}

func (q *ByteChunkQ) ReleaseWrite() {
	atomic.AddInt64(&q.write.Value, q.chunk)
}

func (q *ByteChunkQ) ReleaseWriteLazy() {
	fatomic.LazyStore(&q.write.Value, q.write.Value+q.chunk)
}

func (q *ByteChunkQ) AcquireRead() []byte {
	chunk := q.chunk
	read := q.read.Value
	readTo := read + chunk
	if readTo > q.writeCache.Value {
		q.writeCache.Value = atomic.LoadInt64(&q.write.Value)
		if readTo > q.writeCache.Value {
			q.failedReads.Value++
			return nil
		}
	}
	idx := read & q.mask
	nxt := idx + chunk
	return q.ringBuffer[idx:nxt]
}

func (q *ByteChunkQ) ReleaseRead() {
	atomic.AddInt64(&q.read.Value, q.chunk)
}

func (q *ByteChunkQ) ReleaseReadLazy() {
	fatomic.LazyStore(&q.read.Value, q.read.Value+q.chunk)
}
