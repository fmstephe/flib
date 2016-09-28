// Copyright 2016 Francis Stephens. All rights reserved.
// Use of this source code is governed by a BSD
// license which can be found in LICENSE.txt

package spscq

import (
	"sync/atomic"
	"unsafe"

	"github.com/fmstephe/flib/fmath"
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
	return &PointerQ{ringBuffer: ringBuffer, commonQ: cq}, nil
}

func (q *PointerQ) AcquireRead(bufferSize int64) []unsafe.Pointer {
	readTo := q.read.Value + bufferSize
	if readTo > q.writeCache.Value {
		q.writeCache.Value = atomic.LoadInt64(&q.write.Value)
		if readTo > q.writeCache.Value {
			bufferSize = q.writeCache.Value - q.read.Value
			if bufferSize == 0 {
				q.failedReads.Value++
				ftime.Pause(q.pause)
				return nil
			}
		}
	}
	from := q.read.Value & q.mask
	to := fmath.Min(from+bufferSize, q.size)
	q.readSize.Value = to - from
	return q.ringBuffer[from:to]
}

func (q *PointerQ) AcquireWrite(bufferSize int64) []unsafe.Pointer {
	writeTo := q.write.Value + bufferSize
	readLimit := writeTo - q.size
	if readLimit > q.readCache.Value {
		q.readCache.Value = atomic.LoadInt64(&q.read.Value)
		if readLimit > q.readCache.Value {
			q.failedWrites.Value++
			ftime.Pause(q.pause)
			return nil
		}
	}
	from := q.write.Value & q.mask
	to := fmath.Min(from+bufferSize, q.size)
	q.writeSize.Value = to - from
	return q.ringBuffer[from:to]
}

func (q *PointerQ) WriteSingle(val unsafe.Pointer) bool {
	b := q.writeSingle(val)
	if b {
		atomic.AddInt64(&q.write.Value, 1)
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
		fatomic.LazyStore(&q.write.Value, q.write.Value+1)
	}
	return b
}

func (q *PointerQ) writeSingle(val unsafe.Pointer) bool {
	write := q.write.Value
	readLimit := write - q.size
	if readLimit == q.readCache.Value {
		q.readCache.Value = atomic.LoadInt64(&q.read.Value)
		if readLimit == q.readCache.Value {
			q.failedWrites.Value++
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
		atomic.AddInt64(&q.read.Value, 1)
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
		fatomic.LazyStore(&q.read.Value, q.read.Value+1)
	}
	return val
}

func (q *PointerQ) readSingle() unsafe.Pointer {
	read := q.read.Value
	if read == q.writeCache.Value {
		q.writeCache.Value = atomic.LoadInt64(&q.write.Value)
		if read == q.writeCache.Value {
			q.failedReads.Value++
			ftime.Pause(q.pause)
			return nil
		}
	}
	val := q.ringBuffer[read&q.mask]
	return val
}
