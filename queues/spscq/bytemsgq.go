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
	cq, err := newCommonQ(size, pause)
	if err != nil {
		return nil, err // TODO is that the best error to return?
	}
	return &ByteMsgQ{ringBuffer: ringBuffer, commonQ: cq}, nil
}

func (q *ByteMsgQ) AcquireWrite(bufferSize int64) []byte {
	totalSize := bufferSize + headerSize
	initFrom := q.write.Value & q.mask
	rem := q.size - initFrom
	if rem < totalSize {
		if rem >= headerSize {
			writeHeader(q.ringBuffer, initFrom, -rem)
		}
		atomic.AddInt64(&q.write.Value, rem)
	}
	from, to := q.msgWrite(totalSize)
	if from == to {
		return nil
	}
	writeHeader(q.ringBuffer, from, totalSize)
	return q.ringBuffer[from+headerSize : to]
}

func (q *ByteMsgQ) AcquireRead() []byte {
	rem := q.size - (q.read.Value & q.mask)
	if rem < headerSize {
		atomic.AddInt64(&q.read.Value, rem)
	}
	initFrom := q.read.Value & q.mask
	totalSize := readHeader(q.ringBuffer, initFrom)
	if totalSize < 0 {
		atomic.AddInt64(&q.read.Value, -totalSize)
		initFrom = q.read.Value & q.mask
		totalSize = readHeader(q.ringBuffer, initFrom)
	}
	from, to := q.msgRead(totalSize)
	if from == to {
		return nil
	}
	return q.ringBuffer[from+headerSize : to]
}

func (q *ByteMsgQ) msgWrite(bufferSize int64) (from int64, to int64) {
	writeTo := q.write.Value + bufferSize
	readLimit := writeTo - q.size
	if readLimit > q.readCache.Value {
		q.readCache.Value = atomic.LoadInt64(&q.read.Value)
		if readLimit > q.readCache.Value {
			q.failedWrites.Value++
			ftime.Pause(q.pause)
			return 0, 0
		}
	}
	from = q.write.Value & q.mask
	to = from + bufferSize
	q.writeSize.Value = bufferSize
	return from, to
}

func (q *ByteMsgQ) msgRead(bufferSize int64) (from int64, to int64) {
	readTo := q.read.Value + bufferSize
	if readTo > q.writeCache.Value {
		q.writeCache.Value = atomic.LoadInt64(&q.write.Value)
		if readTo > q.writeCache.Value {
			q.failedReads.Value++
			ftime.Pause(q.pause)
			return 0, 0
		}
	}
	from = q.read.Value & q.mask
	to = from + bufferSize
	q.readSize.Value = bufferSize
	return from, to
}

func writeHeader(buffer []byte, i, val int64) {
	*((*int64)(unsafe.Pointer(&buffer[i]))) = val
}

func readHeader(buffer []byte, i int64) int64 {
	return *((*int64)(unsafe.Pointer(&buffer[i])))
}
