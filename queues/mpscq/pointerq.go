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
	// SingleBlocking Read/Write
	ReadSingleBlocking() unsafe.Pointer
	WriteSingleBlocking(unsafe.Pointer)
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

func (q *PointerQ) WriteSingleBlocking(val unsafe.Pointer) {
	first := false
	for {
		write := atomic.LoadInt64(&q.write.Value)
		if atomic.CompareAndSwapPointer(&q.ringBuffer[write&q.mask], nil, val) {
			println("Success", write)
			atomic.AddInt64(&q.write.Value, 1)
			return
		}
		if !first {
			println("Failed", q.ringBuffer[write&q.mask], atomic.LoadInt64(&q.write.Value))
			first = true
		}
		ftime.Pause(q.pause)
	}
}

func (q *PointerQ) ReadSingleBlocking() unsafe.Pointer {
	read := q.read.Value
	for {
		val := atomic.LoadPointer(&q.ringBuffer[read&q.mask])
		if val != nil {
			atomic.StorePointer(&q.ringBuffer[read&q.mask], nil)
			q.read.Value++
			println("read", q.read.Value)
			return val
		} else {
			q.failedReads.Value++
			ftime.Pause(q.pause)
		}
	}
}
