// Copyright 2016 Francis Stephens. All rights reserved.
// Use of this source code is governed by a BSD
// license which can be found in LICENSE.txt

package mpscq

import (
	"sync/atomic"
	"unsafe"

	"github.com/fmstephe/flib/fsync/padded"
	"github.com/fmstephe/flib/ftime"
)

type PointerQueue interface {
	// Single Read/Write
	ReadSingle() unsafe.Pointer
	WriteSingle(unsafe.Pointer) bool
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

func (q *PointerQ) WriteSingle(val unsafe.Pointer) bool {
	for {
		write := atomic.LoadInt64(&q.write.Value)
		readLimit := write - q.size
		if readLimit == atomic.LoadInt64(&q.readCache.Value) {
			readCache := atomic.LoadInt64(&q.read.Value)
			if !atomic.CompareAndSwapInt64(&q.readCache.Value, q.readCache.Value, readCache) {
				ftime.Pause(q.pause)
				continue
			}
			if readLimit == q.readCache.Value {
				q.failedWrites.Value++
				ftime.Pause(q.pause)
				return false
			}
		}
		if atomic.CompareAndSwapPointer(&q.ringBuffer[write&q.mask], nil, val) {
			atomic.AddInt64(&q.write.Value, 1)
			return true
		}
		ftime.Pause(q.pause)
	}
}

func (q *PointerQ) ReadSingle() unsafe.Pointer {
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
	q.ringBuffer[read&q.mask] = nil
	atomic.AddInt64(&q.read.Value, 1)
	return val
}
