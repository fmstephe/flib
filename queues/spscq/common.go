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
	_readonlyBuffer padded.CacheBuffer
	name            string
	size            int64
	mask            int64
	pause           int64
	_mutableBuffer  padded.CacheBuffer
	offset          int64
	failed          int64
	oppositeCache   int64
	unreleased      int64
	_oppositeBuffer padded.CacheBuffer
	opposite        *int64
	_releasedBuffer padded.CacheBuffer
	released        int64
	_endBuffer      padded.CacheBuffer
}

func (f *mutableFields) acquire(bufferSize int64) (from, to int64) {
	acquireFrom := (f.released - f.offset)
	acquireTo := acquireFrom + bufferSize
	if acquireTo > f.oppositeCache {
		f.oppositeCache = atomic.LoadInt64(f.opposite)
		if acquireTo > f.oppositeCache {
			bufferSize = f.oppositeCache - acquireFrom
			if bufferSize == 0 {
				f.failed++
				ftime.Pause(f.pause)
				return 0, 0
			}
		}
	}
	from = f.released & f.mask
	to = fmath.Min(from+bufferSize, f.size)
	f.unreleased = to - from
	return from, to
}

func (f *mutableFields) release() {
	atomic.AddInt64(&f.released, f.unreleased)
	f.unreleased = 0
}

func (f *mutableFields) releaseLazy() {
	fatomic.LazyStore(&f.released, f.released+f.unreleased)
	f.unreleased = 0
}

func (f *mutableFields) getFailed() int64 {
	return atomic.LoadInt64(&f.failed)
}

func (f *mutableFields) String() string {
	released := f.released
	unreleased := f.unreleased
	failed := f.failed
	cached := f.oppositeCache
	return fmt.Sprintf("{%s, size %d, mask %d, released %d(%d), unreleased %d, failed %d, cached %d(%d) }", f.name, f.size, f.mask, released, released&f.mask, unreleased, failed, cached, cached&f.mask)
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
		write: mutableFields{name: "write", size: size, mask: size - 1, pause: pause},
		read:  mutableFields{name: "read", size: size, mask: size - 1, pause: pause},
	}
	q.write.offset = size
	return q, nil
}

func (q *commonQ) initialise() {
	q.write.opposite = &q.read.released
	q.read.opposite = &q.write.released
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
	return fmt.Sprintf("%s, %s", q.read, q.write)
}

func (q *commonQ) readString() string {
	return q.read.String()
}

func (q *commonQ) writeString() string {
	return q.write.String()
}
