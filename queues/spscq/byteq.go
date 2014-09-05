package spscq

import (
	"github.com/fmstephe/flib/fsync/padded"
	"sync/atomic"
)

type ByteQueue interface {
	// Simple Read/Write
	Read([]byte) bool
	Write([]byte) bool
	//Acquire/Release Read
	AcquireRead(int64) []byte
	ReleaseRead()
	ReleaseReadLazy()
	//Acquire/Release Write
	AcquireWrite(int64) []byte
	ReleaseWrite()
	ReleaseWriteLazy()
}

func NewByteQueue(size int64) (ByteQueue, error) {
	return NewByteQ(size)
}

type ByteQ struct {
	_prebuffer padded.CacheBuffer
	commonQ
	_midbuffer  padded.CacheBuffer
	ringBuffer  []byte
	_postbuffer padded.CacheBuffer
}

func NewByteQ(size int64) (*ByteQ, error) {
	ringBuffer := padded.ByteSlice(int(size))
	cq, err := newCommonQ(size)
	if err != nil {
		return nil, err
	}
	return &ByteQ{ringBuffer: ringBuffer, commonQ: cq}, nil
}

func (q *ByteQ) AcquireWrite(bufferSize int64) []byte {
	from, to := q.acquireWrite(bufferSize)
	return q.ringBuffer[from:to]
}

func (q *ByteQ) AcquireRead(bufferSize int64) []byte {
	from, to := q.acquireRead(bufferSize)
	return q.ringBuffer[from:to]
}

func (q *ByteQ) Write(buffer []byte) bool {
	bufferSize := int64(len(buffer))
	from, to, wrap := q.writeWrappingBuffer(bufferSize)
	if to == 0 {
		return false
	}
	copy(q.ringBuffer[from:to], buffer)
	if wrap != 0 {
		copy(q.ringBuffer[:wrap], buffer[bufferSize-wrap:])
	}
	atomic.AddInt64(&q.write.Value, bufferSize)
	return true
}

func (q *ByteQ) Read(buffer []byte) bool {
	bufferSize := int64(len(buffer))
	from, to, wrap := q.readWrappingBuffer(bufferSize)
	if to == 0 {
		return false
	}
	copy(buffer, q.ringBuffer[from:to])
	if wrap != 0 {
		copy(buffer[bufferSize-wrap:], q.ringBuffer[:wrap])
	}
	atomic.AddInt64(&q.read.Value, bufferSize)
	return true
}
