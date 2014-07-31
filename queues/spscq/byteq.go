package spscq

import (
	"fmt"
	"sync/atomic"

	"github.com/fmstephe/flib/fmath"
	"github.com/fmstephe/flib/fsync/fatomic"
	"github.com/fmstephe/flib/fsync/padded"
)

type ByteQ struct {
	paddedCounters
	ringBuffer  []byte
	size        int64
	mask        int64
	_postbuffer padded.CacheBuffer
}

func NewByteQ(size int64) *ByteQ {
	if !fmath.PowerOfTwo(size) {
		panic(fmt.Sprintf("Size (%d) must be a power of two", size))
	}
	ringBuffer := padded.ByteSlice(int(size))
	q := &ByteQ{ringBuffer: ringBuffer, size: size, mask: size - 1}
	return q
}

func (q *ByteQ) Write(writeBuffer []byte) bool {
	b := q.writeBuffer(writeBuffer)
	if b {
		chunk := int64(len(writeBuffer))
		atomic.AddInt64(&q.write.Value, chunk)
	}
	return b
}

func (q *ByteQ) WriteLazy(writeBuffer []byte) bool {
	b := q.writeBuffer(writeBuffer)
	if b {
		chunk := int64(len(writeBuffer))
		fatomic.LazyStore(&q.write.Value, q.write.Value+chunk)
	}
	return b
}

func (q *ByteQ) writeBuffer(writeBuffer []byte) bool {
	chunk := int64(len(writeBuffer))
	write := q.write.Value
	writeTo := write + chunk
	readLimit := writeTo - q.size
	if readLimit > q.readCache.Value {
		q.readCache.Value = atomic.LoadInt64(&q.read.Value)
		if readLimit > q.readCache.Value {
			q.writeFail.Value++
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
	return true
}

func (q *ByteQ) Read(readBuffer []byte) bool {
	b := q.readBuffer(readBuffer)
	if b {
		chunk := int64(len(readBuffer))
		atomic.AddInt64(&q.read.Value, chunk)
	}
	return b
}

func (q *ByteQ) ReadLazy(readBuffer []byte) bool {
	b := q.readBuffer(readBuffer)
	if b {
		chunk := int64(len(readBuffer))
		fatomic.LazyStore(&q.read.Value, q.read.Value+chunk)
	}
	return b
}

func (q *ByteQ) readBuffer(readBuffer []byte) bool {
	read := q.read.Value
	write := q.writeCache.Value
	if read == write {
		q.writeCache.Value = atomic.LoadInt64(&q.write.Value)
		write = q.writeCache.Value
		if read == write {
			q.readFail.Value++
			return false
		}
	}
	chunk := fmath.Min(write-read, int64(len(readBuffer)))
	idx := read & q.mask
	nxt := idx + chunk
	if nxt <= q.size {
		copy(readBuffer, q.ringBuffer[idx:nxt])
	} else {
		mid := q.size - idx
		copy(readBuffer[:mid], q.ringBuffer[idx:])
		copy(readBuffer[mid:], q.ringBuffer)
	}
	return true
}
