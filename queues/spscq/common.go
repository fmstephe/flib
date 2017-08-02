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
	"github.com/fmstephe/flib/ftime"
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

// PointerQ AcquireRelease methods

func (q *commonQ) acquireRead(bufferSize int64) (from, to int64) {
	readTo := q.read.Value + bufferSize
	if readTo > q.writeCache.Value {
		q.writeCache.Value = atomic.LoadInt64(&q.write.Value)
		if readTo > q.writeCache.Value {
			bufferSize = q.writeCache.Value - q.read.Value
			if bufferSize == 0 {
				q.failedReads.Value++
				ftime.Pause(q.pause)
				return 0, 0
			}
		}
	}
	from = q.read.Value & q.mask
	to = fmath.Min(from+bufferSize, q.size)
	q.readSize.Value = to - from
	return from, to
}

func (q *commonQ) releaseStoredRead() {
	atomic.AddInt64(&q.read.Value, q.readSize.Value)
	q.readSize.Value = 0
}

func (q *commonQ) releaseStoredReadLazy() {
	fatomic.LazyStore(&q.read.Value, q.read.Value+q.readSize.Value)
	q.readSize.Value = 0
}

func (q *commonQ) acquireWrite(bufferSize int64) (from, to int64) {
	offsetWrite := q.write.Value - q.size
	offsetWriteTo := offsetWrite + bufferSize
	if offsetWriteTo > q.readCache.Value {
		q.readCache.Value = atomic.LoadInt64(&q.read.Value)
		if offsetWriteTo > q.readCache.Value {
			bufferSize = q.readCache.Value - offsetWrite
			if bufferSize == 0 {
				q.failedWrites.Value++
				ftime.Pause(q.pause)
				return 0, 0
			}
		}
	}
	from = q.write.Value & q.mask
	to = fmath.Min(from+bufferSize, q.size)
	q.writeSize.Value = to - from
	return from, to
}

func (q *commonQ) releaseStoredWrite() {
	atomic.AddInt64(&q.write.Value, q.writeSize.Value)
	q.writeSize.Value = 0
}

func (q *commonQ) releaseStoredWriteLazy() {
	fatomic.LazyStore(&q.write.Value, q.write.Value+q.writeSize.Value)
	q.writeSize.Value = 0
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
	return fmt.Sprintf("{size %d, mask %d, write %d(%d), writeSize %d, failedWrites %d, readCache %d(%d), read %d(%d), readSize %d, failedReads %d, writeCache %d}", size, mask, write, write&mask, writeSize, failedWrites, readCache, readCache&mask, read, read&mask, readSize, failedReads, writeCache)
}

func (q *commonQ) readString() string {
	size := q.size
	mask := q.mask
	read := q.read.Value
	readSize := q.readSize.Value
	failedReads := q.failedReads.Value
	writeCache := q.writeCache.Value
	return fmt.Sprintf("{read: size %d, mask %d, read %d(%d), readSize %d, failedReads %d, writeCache %d(%d)}", size, mask, read, read&mask, readSize, failedReads, writeCache, read&mask)
}

func (q *commonQ) writeString() string {
	size := q.size
	mask := q.mask
	write := q.write.Value
	writeSize := q.writeSize.Value
	failedWrites := q.failedWrites.Value
	readCache := q.readCache.Value
	return fmt.Sprintf("{write: size %d, mask %d, write %d(%d), writeSize %d, failedWrites %d, readCache %d(%d)}", size, mask, write, write&mask, writeSize, failedWrites, readCache, readCache&mask)
}
