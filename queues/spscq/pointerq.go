// Copyright 2016 Francis Stephens. All rights reserved.
// Use of this source code is governed by a BSD
// license which can be found in LICENSE.txt

package spscq

import (
	"sync/atomic"
	"unsafe"

	"github.com/fmstephe/flib/fsync/fatomic"
	"github.com/fmstephe/flib/fsync/padded"
	"github.com/fmstephe/flib/ftime"
)

type PointerQueue interface {
	// Batch Read/Write
	AcquireRead(int64) []unsafe.Pointer
	ReleaseRead()
	ReleaseReadLazy()
	AcquireWrite(int64) []unsafe.Pointer
	ReleaseWrite()
	ReleaseWriteLazy()
	// Single Read/Write
	ReadSingle() unsafe.Pointer
	WriteSingle(unsafe.Pointer) bool
	ReadSingleBlocking() unsafe.Pointer
	WriteSingleBlocking(unsafe.Pointer)
	ReadSingleLazy() unsafe.Pointer
	WriteSingleLazy(unsafe.Pointer) bool
}

func NewPointerQueue(size, pause int64) (PointerQueue, error) {
	return NewPointerQ(size, pause)
}

type PointerQ struct {
	_prebuffer padded.CacheBuffer
	commonQ
	_midbuffer  padded.CacheBuffer
	ringBuffer  []unsafe.Pointer
	_postbuffer padded.CacheBuffer
}

func NewPointerQ(size, pause int64) (*PointerQ, error) {
	cq, err := newCommonQ(size, pause)
	if err != nil {
		return nil, err
	}
	ringBuffer := padded.PointerSlice(int(size))
	q := &PointerQ{ringBuffer: ringBuffer, commonQ: cq}
	q.initialise()
	return q, nil
}

func (q *PointerQ) AcquireRead(bufferSize int64) []unsafe.Pointer {
	from, to := q.acquireRead(bufferSize)
	return q.ringBuffer[from:to]
}

func (q *PointerQ) ReleaseRead() {
	from := q.read.released.Value & q.mask
	to := from + q.read.unreleased.Value
	for i := from; i < to; i++ {
		q.ringBuffer[i] = nil
	}
	q.releaseStoredRead()
}

func (q *PointerQ) ReleaseReadLazy() {
	from := q.read.released.Value & q.mask
	to := from + q.read.unreleased.Value
	for i := from; i < to; i++ {
		q.ringBuffer[i] = nil
	}
	q.releaseStoredReadLazy()
}

func (q *PointerQ) AcquireWrite(bufferSize int64) []unsafe.Pointer {
	from, to := q.acquireWrite(bufferSize)
	return q.ringBuffer[from:to]
}

func (q *PointerQ) ReleaseWrite() {
	q.releaseStoredWrite()
}

func (q *PointerQ) ReleaseWriteLazy() {
	q.releaseStoredWriteLazy()
}

func (q *PointerQ) WriteSingle(val unsafe.Pointer) bool {
	b := q.writeSingle(val)
	if b {
		atomic.AddInt64(&q.write.released.Value, 1)
	}
	return b
}

func (q *PointerQ) WriteSingleBlocking(val unsafe.Pointer) {
	b := q.WriteSingle(val)
	for !b {
		b = q.WriteSingle(val)
	}
}

func (q *PointerQ) WriteSingleLazy(val unsafe.Pointer) bool {
	b := q.writeSingle(val)
	if b {
		fatomic.LazyStore(&q.write.released.Value, q.write.released.Value+1)
	}
	return b
}

func (q *PointerQ) writeSingle(val unsafe.Pointer) bool {
	write := q.write.released.Value
	readLimit := write - q.size
	if readLimit == q.write.oppositeCache.Value {
		q.write.oppositeCache.Value = atomic.LoadInt64(&q.read.released.Value)
		if readLimit == q.write.oppositeCache.Value {
			q.write.failed.Value++
			ftime.Pause(q.pause)
			return false
		}
	}
	q.ringBuffer[write&q.mask] = val
	return true
}

func (q *PointerQ) ReadSingle() unsafe.Pointer {
	val := q.readSingle()
	if val != nil {
		atomic.AddInt64(&q.read.released.Value, 1)
	}
	return val
}

func (q *PointerQ) ReadSingleBlocking() unsafe.Pointer {
	val := q.ReadSingle()
	for val == nil {
		val = q.ReadSingle()
	}
	return val
}

func (q *PointerQ) ReadSingleLazy() unsafe.Pointer {
	val := q.readSingle()
	if val != nil {
		fatomic.LazyStore(&q.read.released.Value, q.read.released.Value+1)
	}
	return val
}

func (q *PointerQ) readSingle() unsafe.Pointer {
	read := q.read.released.Value
	if read == q.read.oppositeCache.Value {
		q.read.oppositeCache.Value = atomic.LoadInt64(&q.write.released.Value)
		if read == q.read.oppositeCache.Value {
			q.read.failed.Value++
			ftime.Pause(q.pause)
			return nil
		}
	}
	val := q.ringBuffer[read&q.mask]
	q.ringBuffer[read&q.mask] = nil
	return val
}
