// Copyright 2016 Francis Stephens. All rights reserved.
// Use of this source code is governed by a BSD
// license which can be found in LICENSE.txt

package spscq

import (
	"sync/atomic"
	"unsafe"

	"github.com/fmstephe/flib/fsync/padded"
	"github.com/fmstephe/flib/ftime"
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
	cq, err := newPointerCommonQ(size, pause)
	if err != nil {
		return nil, err // TODO is that the best error to return?
	}
	return &ByteMsgQ{ringBuffer: ringBuffer, commonQ: cq}, nil
}

func (q *ByteMsgQ) AcquireWrite(bufferSize int64) []byte {
	totalSize := bufferSize + headerSize
	initFrom := q.write.released & q.mask
	rem := q.size - initFrom
	if rem < totalSize {
		totalSize += rem
	}
	from, to := q.msg_write_acquire(totalSize)
	if from == to {
		return nil
	}
	if rem >= headerSize {
		writeHeader(q.ringBuffer, initFrom, -rem)
	}
	writeHeader(q.ringBuffer, from, totalSize)
	return q.ringBuffer[from+headerSize : to]
}

func (q *ByteMsgQ) ReleaseWrite() {
	q.write.release()
}

func (q *ByteMsgQ) ReleaseWriteLazy() {
	q.write.releaseLazy()
}

func (q *ByteMsgQ) AcquireRead() []byte {
	rem := q.size - (q.read.released & q.mask)
	if rem < headerSize {
		atomic.AddInt64(&q.read.released, rem)
	}
	initFrom := q.read.released & q.mask
	totalSize := readHeader(q.ringBuffer, initFrom)
	if totalSize < 0 {
		atomic.AddInt64(&q.read.released, -totalSize)
		initFrom = q.read.released & q.mask
		totalSize = readHeader(q.ringBuffer, initFrom)
	}
	from, to := q.msg_read_acquire(totalSize)
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

func (q *ByteMsgQ) msg_write_acquire(bufferSize int64) (from int64, to int64) {
	acquireFrom := q.write.released - q.write.offset
	acquireTo := acquireFrom + bufferSize
	if acquireTo > q.write.oppositeCache {
		q.write.oppositeCache = atomic.LoadInt64(q.write.opposite)
		if acquireTo > q.write.oppositeCache {
			q.write.failed++
			ftime.Pause(q.pause)
			return 0, 0
		}
	}
	from = q.write.released & q.write.mask
	to = from + bufferSize
	q.write.unreleased = bufferSize
	return from, to
}

func (q *ByteMsgQ) msg_read_acquire(bufferSize int64) (from int64, to int64) {
	acquireFrom := q.read.released - q.read.offset
	acquireTo := acquireFrom + bufferSize
	if acquireTo > q.read.oppositeCache {
		q.read.oppositeCache = atomic.LoadInt64(q.read.opposite)
		if acquireTo > q.read.oppositeCache {
			q.read.failed++
			ftime.Pause(q.pause)
			return 0, 0
		}
	}
	from = q.read.released & q.read.mask
	to = from + bufferSize
	q.read.unreleased = bufferSize
	return from, to
}

func writeHeader(buffer []byte, i, val int64) {
	*((*int64)(unsafe.Pointer(&buffer[i]))) = val
}

func readHeader(buffer []byte, i int64) int64 {
	return *((*int64)(unsafe.Pointer(&buffer[i])))
}
