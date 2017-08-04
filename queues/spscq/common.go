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
	// TODO Remove these
	size  int64
	mask  int64
	pause int64
	// Writer fields
	write mutableFields
	// Reader fields
	read mutableFields
}

type mutableFields struct {
	// TODO have a hard think about padding in this struct
	// Readonly Fields
	size  int64
	mask  int64
	pause int64
	// Shared Mutable Field
	released padded.Int64
	// Private Mutable Fields
	opposite      *int64
	offset        padded.Int64
	unreleased    padded.Int64
	failed        padded.Int64
	oppositeCache padded.Int64
}

func (f *mutableFields) acquire(bufferSize int64) (from, to int64) {
	acquireFrom := (f.released.Value - f.offset.Value)
	acquireTo := acquireFrom + bufferSize
	if acquireTo > f.oppositeCache.Value {
		f.oppositeCache.Value = atomic.LoadInt64(f.opposite)
		if acquireTo > f.oppositeCache.Value {
			bufferSize = f.oppositeCache.Value - acquireFrom
			if bufferSize == 0 {
				f.failed.Value++
				ftime.Pause(f.pause)
				return 0, 0
			}
		}
	}
	from = f.released.Value & f.mask
	to = fmath.Min(from+bufferSize, f.size)
	f.unreleased.Value = to - from
	return from, to
}

func (f *mutableFields) release() {
	atomic.AddInt64(&f.released.Value, f.unreleased.Value)
	f.unreleased.Value = 0
}

func (f *mutableFields) releaseLazy() {
	fatomic.LazyStore(&f.released.Value, f.released.Value+f.unreleased.Value)
	f.unreleased.Value = 0
}

func (f *mutableFields) getFailed() int64 {
	return atomic.LoadInt64(&f.failed.Value)
}

func newCommonQ(size, pause int64) (commonQ, error) {
	if !fmath.PowerOfTwo(size) {
		return commonQ{}, errors.New(fmt.Sprintf("Size (%d) must be a power of two", size))
	}
	if size > maxSize {
		return commonQ{}, errors.New(fmt.Sprintf("Size (%d) must be less than %d", size, maxSize))
	}
	q := commonQ{
		size:  size,
		mask:  size - 1,
		pause: pause,
		write: mutableFields{size: size, mask: size - 1, pause: pause},
		read:  mutableFields{size: size, mask: size - 1, pause: pause},
	}
	q.write.offset.Value = size
	return q, nil
}

func (q *commonQ) initialise() {
	q.write.opposite = &q.read.released.Value
	q.read.opposite = &q.write.released.Value
}

// PointerQ AcquireRelease methods

func (q *commonQ) acquireRead(bufferSize int64) (from, to int64) {
	return q.read.acquire(bufferSize)
}

func (q *commonQ) acquireWrite(bufferSize int64) (from, to int64) {
	return q.write.acquire(bufferSize)
}

func (q *commonQ) releaseStoredRead() {
	q.read.release()
}

func (q *commonQ) releaseStoredReadLazy() {
	q.read.releaseLazy()
}

func (q *commonQ) releaseStoredWrite() {
	q.write.release()
}

func (q *commonQ) releaseStoredWriteLazy() {
	q.write.releaseLazy()
}

func (q *commonQ) FailedReads() int64 {
	return q.read.getFailed()
}

func (q *commonQ) FailedWrites() int64 {
	return q.write.getFailed()
}

func (q *commonQ) String() string {
	size := q.size
	mask := q.mask
	write := q.write.released.Value
	writeUnreleased := q.write.unreleased.Value
	writeFailed := q.write.failed.Value
	readCache := q.write.oppositeCache.Value
	read := q.read.released.Value
	readUnreleased := q.read.unreleased.Value
	readFailed := q.read.failed.Value
	writeCache := q.read.oppositeCache.Value
	return fmt.Sprintf("{size %d, mask %d, write %d(%d), writeUnreleased %d, writeFailed %d, readCache %d(%d), read %d(%d), readUnreleased %d, readFailed %d, writeCache %d}", size, mask, write, write&mask, writeUnreleased, writeFailed, readCache, readCache&mask, read, read&mask, readUnreleased, readFailed, writeCache)
}

func (q *commonQ) readString() string {
	size := q.size
	mask := q.mask
	read := q.read.released.Value
	readUnreleased := q.read.unreleased.Value
	readFailed := q.read.failed.Value
	writeCache := q.read.oppositeCache.Value
	return fmt.Sprintf("{read: size %d, mask %d, read %d(%d), readUnreleased %d, readfailed %d, writeCache %d(%d)}", size, mask, read, read&mask, readUnreleased, readFailed, writeCache, read&mask)
}

func (q *commonQ) writeString() string {
	size := q.size
	mask := q.mask
	write := q.write.released.Value
	writeUnreleased := q.write.unreleased.Value
	writeFailed := q.write.failed.Value
	readCache := q.write.oppositeCache.Value
	return fmt.Sprintf("{write: size %d, mask %d, write %d(%d), writeUnreleased %d, writeFailed %d, readCache %d(%d)}", size, mask, write, write&mask, writeUnreleased, writeFailed, readCache, readCache&mask)
}
