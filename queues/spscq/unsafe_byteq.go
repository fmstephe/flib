package spscq

import (
	"fmt"
	"github.com/fmstephe/flib/fsync/fatomic"
	"github.com/fmstephe/flib/fsync/padded"
	"sync/atomic"
)

type UnsafeByteQ struct {
	_1         padded.Int64
	read       padded.Int64
	readCache  padded.Int64
	write      padded.Int64
	writeCache padded.Int64
	_2         padded.Int64
	// Read only
	ringBuffer []byte
	size       int64
	mask       int64
	_3         padded.Int64
}

func NewUnsafeByteQ(size int64) *UnsafeByteQ {
	if !powerOfTwo(size) {
		panic(fmt.Sprintf("Size (%d) must be a power of two", size))
	}
	ringBuffer := padded.ByteSlice(int(size))
	q := &UnsafeByteQ{ringBuffer: ringBuffer, size: size, mask: size - 1}
	return q
}

func (q *UnsafeByteQ) Write(writeBuffer []byte) bool {
	chunk := int64(len(writeBuffer))
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
	if nxt <= q.size {
		copy(q.ringBuffer[idx:nxt], writeBuffer)
	} else {
		mid := q.size - idx
		copy(q.ringBuffer[idx:], writeBuffer[:mid])
		copy(q.ringBuffer, writeBuffer[mid:])
	}
	fatomic.LazyStore(&q.write.Value, q.write.Value+chunk)
	return true
}

func (q *UnsafeByteQ) Read(readBuffer []byte) bool {
	read := q.read.Value
	write := q.writeCache.Value
	if read == write {
		q.writeCache.Value = atomic.LoadInt64(&q.write.Value)
		write = q.writeCache.Value
		if read == write {
			return false
		}
	}
	chunk := min(write-read, int64(len(readBuffer)))
	idx := read & q.mask
	nxt := idx + chunk
	if nxt <= q.size {
		copy(readBuffer, q.ringBuffer[idx:nxt])
	} else {
		mid := q.size - idx
		copy(readBuffer[:mid], q.ringBuffer[idx:])
		copy(readBuffer[mid:], q.ringBuffer)
	}
	fatomic.LazyStore(&q.read.Value, q.read.Value+chunk)
	return true
}
