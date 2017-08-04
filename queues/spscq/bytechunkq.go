// Copyright 2016 Francis Stephens. All rights reserved.
// Use of this source code is governed by a BSD
// license which can be found in LICENSE.txt

package spscq

import (
	"errors"
	"fmt"
	"sync/atomic"

	"github.com/fmstephe/flib/fsync/fatomic"
	"github.com/fmstephe/flib/fsync/padded"
	"github.com/fmstephe/flib/ftime"
)

type ByteChunkQueue interface {
	//Acquire/Release Read
	AcquireRead() []byte
	ReleaseRead()
	ReleaseReadLazy()
	//Acquire/Release Write
	AcquireWrite() []byte
	ReleaseWrite()
	ReleaseWriteLazy()
}

func NewByteChunkQueue(size, pause, chunk int64) (ByteChunkQueue, error) {
	return NewByteChunkQ(size, pause, chunk)
}

type ByteChunkQ struct {
	_prebuffer padded.CacheBuffer
	commonQ
	_midbuffer  padded.CacheBuffer
	ringBuffer  []byte
	chunk       int64
	_postbuffer padded.CacheBuffer
}

func NewByteChunkQ(size, pause, chunk int64) (*ByteChunkQ, error) {
	if size%chunk != 0 {
		return nil, errors.New(fmt.Sprintf("Size must divide by chunk, (size) %d rem (chunk) %d = %d", size, chunk, size%chunk))
	}
	ringBuffer := padded.ByteSlice(int(size))
	cq, err := newCommonQ(size, pause)
	if err != nil {
		return nil, err // TODO is that the best error to return?
	}
	return &ByteChunkQ{ringBuffer: ringBuffer, commonQ: cq, chunk: chunk}, nil
}

func (q *ByteChunkQ) AcquireWrite() []byte {
	chunk := q.chunk
	write := q.write.released.Value
	writeTo := write + chunk
	readLimit := writeTo - q.size
	if readLimit > q.write.oppositeCache.Value {
		q.write.oppositeCache.Value = atomic.LoadInt64(&q.read.released.Value)
		if readLimit > q.write.oppositeCache.Value {
			q.write.failed.Value++
			ftime.Pause(q.pause)
			return nil
		}
	}
	idx := write & q.mask
	nxt := idx + chunk
	return q.ringBuffer[idx:nxt]
}

func (q *ByteChunkQ) ReleaseWrite() {
	atomic.AddInt64(&q.write.released.Value, q.chunk)
}

func (q *ByteChunkQ) ReleaseWriteLazy() {
	fatomic.LazyStore(&q.write.released.Value, q.write.released.Value+q.chunk)
}

func (q *ByteChunkQ) AcquireRead() []byte {
	chunk := q.chunk
	read := q.read.released.Value
	readTo := read + chunk
	if readTo > q.read.oppositeCache.Value {
		q.read.oppositeCache.Value = atomic.LoadInt64(&q.write.released.Value)
		if readTo > q.read.oppositeCache.Value {
			q.read.failed.Value++
			ftime.Pause(q.pause)
			return nil
		}
	}
	idx := read & q.mask
	nxt := idx + chunk
	return q.ringBuffer[idx:nxt]
}

func (q *ByteChunkQ) ReleaseRead() {
	atomic.AddInt64(&q.read.released.Value, q.chunk)
}

func (q *ByteChunkQ) ReleaseReadLazy() {
	fatomic.LazyStore(&q.read.released.Value, q.read.released.Value+q.chunk)
}
