package spscq

import (
	"fmt"
	"sync/atomic"

	"github.com/fmstephe/flib/fmath"
	"github.com/fmstephe/flib/fsync/fatomic"
	"github.com/fmstephe/flib/fsync/padded"
)

type ByteChunkQ struct {
	paddedCounters
	ringBuffer  []byte
	size        int64
	chunk       int64
	mask        int64
	_postbuffer padded.CacheBuffer
}

func NewByteChunkQ(size int64, chunk int64) *ByteChunkQ {
	if !fmath.PowerOfTwo(size) {
		panic(fmt.Sprintf("Size must be a power of two, size = %d", size))
	}
	if size%chunk != 0 {
		panic(fmt.Sprintf("Size must be neatly divisible by chunk, (size) %d rem (chunk) %d = %d", size, chunk, size%chunk))
	}
	ringBuffer := padded.ByteSlice(int(size))
	q := &ByteChunkQ{ringBuffer: ringBuffer, size: size, chunk: chunk, mask: size - 1}
	return q
}

func (q *ByteChunkQ) WriteBuffer() []byte {
	chunk := q.chunk
	write := q.write.Value
	writeTo := write + chunk
	readLimit := writeTo - q.size
	if readLimit > q.readCache.Value {
		q.readCache.Value = atomic.LoadInt64(&q.read.Value)
		if readLimit > q.readCache.Value {
			q.writeFail.Value++
			return nil
		}
	}
	idx := write & q.mask
	nxt := idx + chunk
	return q.ringBuffer[idx:nxt]
}

func (q *ByteChunkQ) CommitWrite() {
	atomic.AddInt64(&q.write.Value, q.chunk)
}

func (q *ByteChunkQ) CommitWriteLazy() {
	fatomic.LazyStore(&q.write.Value, q.write.Value+q.chunk)
}

func (q *ByteChunkQ) ReadBuffer() []byte {
	chunk := q.chunk
	read := q.read.Value
	readTo := read + chunk
	if readTo > q.writeCache.Value {
		q.writeCache.Value = atomic.LoadInt64(&q.write.Value)
		if readTo > q.writeCache.Value {
			q.readFail.Value++
			return nil
		}
	}
	idx := read & q.mask
	nxt := idx + chunk
	return q.ringBuffer[idx:nxt]
}

func (q *ByteChunkQ) CommitRead() {
	atomic.AddInt64(&q.read.Value, q.chunk)
}

func (q *ByteChunkQ) CommitReadLazy() {
	fatomic.LazyStore(&q.read.Value, q.read.Value+q.chunk)
}
