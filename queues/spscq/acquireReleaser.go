// Copyright 2016 Francis Stephens. All rights reserved.
// Use of this source code is governed by a BSD
// license which can be found in LICENSE.txt

package spscq

import (
	"fmt"
	"sync/atomic"

	"github.com/fmstephe/flib/fmath"
	"github.com/fmstephe/flib/fsync/fatomic"
	"github.com/fmstephe/flib/fsync/padded"
	"github.com/fmstephe/flib/ftime"
)

type acquireReleaser struct {
	_readonlyBuffer padded.CacheBuffer
	name            string
	size            int64
	mask            int64
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
	released := f.released
	unreleased := f.unreleased
	failed := f.failed
	cached := f.oppositeCache
	return fmt.Sprintf("{%s, size %d, mask %d, released %d(%d), unreleased %d, failed %d, cached %d(%d) }", f.name, f.size, f.mask, released, released&f.mask, unreleased, failed, cached, cached&f.mask)
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

func (f *acquireReleaser) pointerq_release() {
	atomic.AddInt64(&f.released, f.unreleased)
}

func (f *acquireReleaser) pointerq_releaseLazy() {
	fatomic.LazyStore(&f.released, f.released+f.unreleased)
}
