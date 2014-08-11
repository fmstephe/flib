package spscq

import (
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
	from, to := q.writeBuffer(bufferSize)
	return q.ringBuffer[from:to]
}

func (q *ByteQ) ReadBuffer(bufferSize int64) []byte {
	from, to := q.readBuffer(bufferSize)
	return q.ringBuffer[from:to]
}
