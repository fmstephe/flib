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
	// Simple Read/Write
	Read([]unsafe.Pointer) []unsafe.Pointer
	Write([]unsafe.Pointer) bool
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

func (q *PointerQ) Read(buffer []unsafe.Pointer) []unsafe.Pointer {
	from, to := q.acquireRead(int64(len(buffer)))
	bufferSize := to - from
	if bufferSize == 0 {
		return nil
	}
	copy(buffer, q.ringBuffer[from:to])
	atomic.AddInt64(&q.read.Value, bufferSize)
	return buffer[:bufferSize]
}

func (q *PointerQ) Write(buffer []unsafe.Pointer) bool {
	bufferSize := int64(len(buffer))
	from, to := q.acquireWrite(bufferSize)
	if to == 0 {
		return false
	}
	copy(q.ringBuffer[from:to], buffer)
	atomic.AddInt64(&q.write.Value, bufferSize)
	return true
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
