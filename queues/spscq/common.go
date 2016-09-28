// Copyright 2016 Francis Stephens. All rights reserved.
// Use of this source code is governed by a BSD
// license which can be found in LICENSE.txt

package spscq

import (
	"errors"
	"fmt"
	"sync/atomic"

	"github.com/fmstephe/flib/fmath"
	"github.com/fmstephe/flib/fsync/fatomic"
	"github.com/fmstephe/flib/fsync/padded"
)

const maxSize = 1 << 41

type commonQ struct {
	// Readonly Fields
	size  int64
	mask  int64
	pause int64
	// Writer fields
	write        padded.Int64
	writeSize    padded.Int64
	failedWrites padded.Int64
	readCache    padded.Int64
	// Reader fields
	read        padded.Int64
	readSize    padded.Int64
	failedReads padded.Int64
	writeCache  padded.Int64
}

func newCommonQ(size, pause int64) (commonQ, error) {
	var cq commonQ
	if !fmath.PowerOfTwo(size) {
		return cq, errors.New(fmt.Sprintf("Size (%d) must be a power of two", size))
	}
	if size > maxSize {
		return cq, errors.New(fmt.Sprintf("Size (%d) must be less than %d", size, maxSize))
	}
	return commonQ{size: size, mask: size - 1, pause: pause}, nil
}

func (q *commonQ) ReleaseWrite() {
	atomic.AddInt64(&q.write.Value, q.writeSize.Value)
	q.writeSize.Value = 0
}

func (q *commonQ) ReleaseWriteLazy() {
	fatomic.LazyStore(&q.write.Value, q.write.Value+q.writeSize.Value)
	q.writeSize.Value = 0
}

func (q *commonQ) ReleaseRead() {
	atomic.AddInt64(&q.read.Value, q.readSize.Value)
	q.readSize.Value = 0
}

func (q *commonQ) ReleaseReadLazy() {
	fatomic.LazyStore(&q.read.Value, q.read.Value+q.readSize.Value)
	q.readSize.Value = 0
}

func (q *commonQ) FailedWrites() int64 {
	return atomic.LoadInt64(&q.failedWrites.Value)
}

func (q *commonQ) FailedReads() int64 {
	return atomic.LoadInt64(&q.failedReads.Value)
}

func (q *commonQ) String() string {
	size := q.size
	mask := q.mask
	write := q.write.Value
	writeSize := q.writeSize.Value
	failedWrites := q.failedWrites.Value
	readCache := q.readCache.Value
	read := q.read.Value
	readSize := q.readSize.Value
	failedReads := q.failedReads.Value
	writeCache := q.writeCache.Value
	return fmt.Sprintf("{Size %d, mask %d, write %d, writeSize %d, failedWrites %d, readCache %d, read %d, readSize %d, failedReads %d, writeCache %d}", size, mask, write, writeSize, failedWrites, readCache, read, readSize, failedReads, writeCache)
}
