// Copyright 2016 Francis Stephens. All rights reserved.
// Use of this source code is governed by a BSD
// license which can be found in LICENSE.txt

package spscq

import (
	"errors"
	"fmt"

	"github.com/fmstephe/flib/fsync/padded"
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
	chunk       int64
	_midbuffer  padded.CacheBuffer
	ringBuffer  []byte
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
	q := &ByteChunkQ{commonQ: cq, chunk: chunk, ringBuffer: ringBuffer}
	q.initialise()
	return q, nil
}

func (q *ByteChunkQ) AcquireWrite() []byte {
	from, to := q.write.acquireExactly(q.chunk)
	return q.ringBuffer[from:to]
}

func (q *ByteChunkQ) ReleaseWrite() {
	q.write.release()
}

func (q *ByteChunkQ) ReleaseWriteLazy() {
	q.write.releaseLazy()
}

func (q *ByteChunkQ) AcquireRead() []byte {
	from, to := q.read.acquireExactly(q.chunk)
	return q.ringBuffer[from:to]
}

func (q *ByteChunkQ) ReleaseRead() {
	q.read.release()
}

func (q *ByteChunkQ) ReleaseReadLazy() {
	q.read.releaseLazy()
}
