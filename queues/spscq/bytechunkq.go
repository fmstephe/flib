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
	write := q.write.released
	writeTo := write + chunk
	readLimit := writeTo - q.size
	if readLimit > q.write.oppositeCache {
		q.write.oppositeCache = atomic.LoadInt64(&q.read.released)
		if readLimit > q.write.oppositeCache {
			q.write.failed++
			ftime.Pause(q.pause)
			return nil
		}
	}
	idx := write & q.mask
	nxt := idx + chunk
	return q.ringBuffer[idx:nxt]
}

func (q *ByteChunkQ) ReleaseWrite() {
	atomic.AddInt64(&q.write.released, q.chunk)
}

func (q *ByteChunkQ) ReleaseWriteLazy() {
	fatomic.LazyStore(&q.write.released, q.write.released+q.chunk)
}

func (q *ByteChunkQ) AcquireRead() []byte {
	chunk := q.chunk
	read := q.read.released
	readTo := read + chunk
	if readTo > q.read.oppositeCache {
		q.read.oppositeCache = atomic.LoadInt64(&q.write.released)
		if readTo > q.read.oppositeCache {
			q.read.failed++
			ftime.Pause(q.pause)
			return nil
		}
	}
	idx := read & q.mask
	nxt := idx + chunk
	return q.ringBuffer[idx:nxt]
}

func (q *ByteChunkQ) ReleaseRead() {
	atomic.AddInt64(&q.read.released, q.chunk)
}

func (q *ByteChunkQ) ReleaseReadLazy() {
	fatomic.LazyStore(&q.read.released, q.read.released+q.chunk)
}
