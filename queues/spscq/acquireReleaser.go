// Copyright 2016 Francis Stephens. All rights reserved.
// Use of this source code is governed by a BSD
// license which can be found in LICENSE.txt

package spscq

import (
	"fmt"
	"sync/atomic"

	"github.com/fmstephe/flib/fmath"
	"github.com/fmstephe/flib/fstrconv"
	"github.com/fmstephe/flib/fsync/fatomic"
	"github.com/fmstephe/flib/fsync/padded"
	"github.com/fmstephe/flib/ftime"
)

type acquireReleaser struct {
	_readonlyBuffer padded.CacheBuffer
	name            string
	size            int64
	mask            int64
	chunk           int64
	pause           int64
	offset          int64
	_mutableBuffer  padded.CacheBuffer
	failed          int64
	oppositeCache   int64
	unreleased      int64
	_oppositeBuffer padded.CacheBuffer
	opposite        *int64
	_releasedBuffer padded.CacheBuffer
	released        int64
	_endBuffer      padded.CacheBuffer
}

func (f *acquireReleaser) getFailed() int64 {
	return atomic.LoadInt64(&f.failed)
}

func (f *acquireReleaser) String() string {
	size := fstrconv.ItoaComma(f.size)
	mask := fstrconv.ItoaComma(f.mask)
	released := fstrconv.ItoaComma(f.released)
	maskedReleased := fstrconv.ItoaComma(f.released & f.mask)
	unreleased := fstrconv.ItoaComma(f.unreleased)
	failed := fstrconv.ItoaComma(f.failed)
	cached := fstrconv.ItoaComma(f.oppositeCache)
	maskedCached := fstrconv.ItoaComma(f.oppositeCache & f.mask)
	offset := fstrconv.ItoaComma(f.offset)
	return fmt.Sprintf("{%s, size %s, mask %s, released %s(%s), unreleased %s, failed %s, cached %s(%s), offset %s }", f.name, size, mask, released, maskedReleased, unreleased, failed, cached, maskedCached, offset)
}

func (f *acquireReleaser) pointerq_acquire(bufferSize int64) (from, to int64) {
	acquireFrom := (f.released - f.offset)
	acquireTo := acquireFrom + bufferSize
	if acquireTo > f.oppositeCache {
		f.oppositeCache = atomic.LoadInt64(f.opposite)
		if acquireTo > f.oppositeCache {
			bufferSize = f.oppositeCache - acquireFrom
			if bufferSize == 0 {
				f.failed++
				f.unreleased = 0
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

func (f *acquireReleaser) bytechunkq_acquire() (from, to int64) {
	bufferSize := f.chunk
	acquireFrom := (f.released - f.offset)
	acquireTo := acquireFrom + bufferSize
	if acquireTo > f.oppositeCache {
		f.oppositeCache = atomic.LoadInt64(f.opposite)
		if acquireTo > f.oppositeCache {
			f.failed++
			ftime.Pause(f.pause)
			return 0, 0
		}
	}
	from = f.released & f.mask
	to = from + bufferSize
	f.unreleased = bufferSize
	return from, to
}

func (f *acquireReleaser) release() {
	atomic.AddInt64(&f.released, f.unreleased)
	f.unreleased = 0
}

func (f *acquireReleaser) releaseLazy() {
	fatomic.LazyStore(&f.released, f.released+f.unreleased)
	f.unreleased = 0
}
