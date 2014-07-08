package spscq

import (
	"fmt"
	"sync/atomic"

	"github.com/fmstephe/flib/fsync/padded"
)

type ByteChunkQ struct {
	_1         padded.CacheBuffer
	read       padded.Int64
	writeCache padded.Int64
	_2         padded.CacheBuffer
	write      padded.Int64
	readCache  padded.Int64
	_3         padded.CacheBuffer
	// Read only
	ringBuffer  []byte
	readBuffer  []byte
	writeBuffer []byte
	size        int64
	chunk       int64
	mask        int64
	_4          padded.CacheBuffer
}

func NewByteChunkQ(size int64, chunk int64) *ByteChunkQ {
	if !powerOfTwo(size) {
		panic(fmt.Sprintf("Size must be a power of two, size = %d", size))
	}
	if size%chunk != 0 {
		panic(fmt.Sprintf("Size must be neatly divisible by chunk, (size) %d rem (chunk) %d = %d", size, chunk, size%chunk))
	}
	ringBuffer := padded.ByteSlice(int(size))
	readBuffer := padded.ByteSlice(int(chunk))
	writeBuffer := padded.ByteSlice(int(chunk))
	q := &ByteChunkQ{ringBuffer: ringBuffer, readBuffer: readBuffer, writeBuffer: writeBuffer, size: size, chunk: chunk, mask: size - 1}
	return q
}

func (q *ByteChunkQ) ReadBuffer() []byte {
	return q.readBuffer
}

func (q *ByteChunkQ) Write() bool {
	chunk := q.chunk
	write := q.write.Value
	writeTo := write + chunk
	readLimit := writeTo - q.size
	if readLimit > q.readCache.Value {
		q.readCache.Value = atomic.LoadInt64(&q.read.Value)
		if readLimit > q.readCache.Value {
			return false
		}
	}
	idx := write & q.mask
	nxt := idx + chunk
	copy(q.ringBuffer[idx:nxt], q.writeBuffer)
	atomic.AddInt64(&q.write.Value, chunk)
	return true
}

func (q *ByteChunkQ) WriteBuffer() []byte {
	return q.writeBuffer
}

func (q *ByteChunkQ) Read() bool {
	chunk := q.chunk
	read := q.read.Value
	readTo := read + chunk
	if readTo > q.writeCache.Value {
		q.writeCache.Value = atomic.LoadInt64(&q.write.Value)
		if readTo > q.writeCache.Value {
			return false
		}
	}
	idx := read & q.mask
	nxt := idx + chunk
	copy(q.readBuffer, q.ringBuffer[idx:nxt])
	atomic.AddInt64(&q.read.Value, chunk)
	return true
}
