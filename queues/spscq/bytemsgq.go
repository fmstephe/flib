// Copyright 2016 Francis Stephens. All rights reserved.
// Use of this source code is governed by a BSD
// license which can be found in LICENSE.txt

package spscq

import (
	"unsafe"

	"github.com/fmstephe/flib/fsync/padded"
)

const (
	headerSize = 8
)

type ByteMsgQueue interface {
	//Acquire/Release Read
	AcquireRead() []byte
	ReleaseRead()
	ReleaseReadLazy()
	//Acquire/Release Write
	AcquireWrite(int64) []byte
	ReleaseWrite()
	ReleaseWriteLazy()
}

func NewByteMsgQueue(size, pause int64) (ByteMsgQueue, error) {
	return NewByteMsgQ(size, pause)
}

type ByteMsgQ struct {
	_prebuffer padded.CacheBuffer
	commonQ
	_midbuffer  padded.CacheBuffer
	ringBuffer  []byte
	_postbuffer padded.CacheBuffer
}

func NewByteMsgQ(size, pause int64) (*ByteMsgQ, error) {
	// TODO there is an effective minimum queue size - should be enforced
	ringBuffer := padded.ByteSlice(int(size))
	cq, err := newCommonQ(size, pause)
	if err != nil {
		return nil, err // TODO is that the best error to return?
	}
	return &ByteMsgQ{ringBuffer: ringBuffer, commonQ: cq}, nil
}

func (q *ByteMsgQ) AcquireWrite(bufferSize int64) []byte {
	msgSize := bufferSize + headerSize
	totalSize := msgSize
	rem := q.size - (q.write.released & q.mask)
	if rem < totalSize {
		totalSize += rem
	}
	from, to := q.write.acquireExactly(totalSize)
	if from == to {
		return nil
	}
	switch {
	case rem > msgSize:
		writeHeader(q.ringBuffer, from, msgSize)
		return q.ringBuffer[from+headerSize : to]
	case rem >= headerSize:
		writeHeader(q.ringBuffer, from, -rem)
		writeHeader(q.ringBuffer, 0, msgSize)
		return q.ringBuffer[headerSize:to]
	case rem > msgSize && rem < headerSize:
		writeHeader(q.ringBuffer, 0, msgSize)
		return q.ringBuffer[headerSize:to]
	default:
		panic("unreachable")
	}
}

func (q *ByteMsgQ) ReleaseWrite() {
	q.write.release()
}

func (q *ByteMsgQ) ReleaseWriteLazy() {
	q.write.releaseLazy()
}

func (q *ByteMsgQ) AcquireRead() []byte {
	wasteOffset := int64(0)
	from := q.read.released & q.mask
	rem := q.size - from
	if rem < headerSize {
		wasteOffset = rem
		from = 0
	}
	msgSize := readHeader(q.ringBuffer, from)
	if msgSize < 0 {
		wasteOffset = rem
		from = 0
		msgSize = readHeader(q.ringBuffer, 0)
	}
	from, to := q.read.acquireExactly(msgSize + wasteOffset)
	if from == to {
		return nil
	}
	return q.ringBuffer[from+headerSize : to]
}

func (q *ByteMsgQ) ReleaseRead() {
	q.read.release()
}

func (q *ByteMsgQ) ReleaseReadLazy() {
	q.read.releaseLazy()
}

func writeHeader(buffer []byte, i, val int64) {
	*((*int64)(unsafe.Pointer(&buffer[i]))) = val
}

func readHeader(buffer []byte, i int64) int64 {
	return *((*int64)(unsafe.Pointer(&buffer[i])))
}
