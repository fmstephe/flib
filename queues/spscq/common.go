package spscq

import (
	"errors"
	"fmt"
	"github.com/fmstephe/flib/fmath"
	"github.com/fmstephe/flib/fsync/fatomic"
	"github.com/fmstephe/flib/fsync/padded"
	"github.com/fmstephe/flib/ftime"
	"sync/atomic"
)

const maxSize = 1 << 41

type commonQ struct {
	// Readonly Fields
	size       int64
	mask       int64
	pause      int64
	_ropadding padded.CacheBuffer
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

func (q *commonQ) acquireWrite(bufferSize int64) (from int64, to int64) {
	write := q.write.Value
	from = write & q.mask
	bufferSize = fmath.Min(bufferSize, q.size-from)
	writeTo := write + bufferSize
	readLimit := writeTo - q.size
	to = from + bufferSize
	if readLimit > q.readCache.Value {
		q.readCache.Value = atomic.LoadInt64(&q.read.Value)
		if readLimit > q.readCache.Value {
			to = q.readCache.Value & q.mask
		}
	}
	if from == to {
		q.failedWrites.Value++
		ftime.Pause(q.pause)
	}
	q.writeSize.Value = to - from
	return from, to
}

func (q *commonQ) ReleaseWrite() {
	atomic.AddInt64(&q.write.Value, q.writeSize.Value)
	q.writeSize.Value = 0
}

func (q *commonQ) ReleaseWriteLazy() {
	fatomic.LazyStore(&q.write.Value, q.write.Value+q.writeSize.Value)
	q.writeSize.Value = 0
}

func (q *commonQ) acquireRead(bufferSize int64) (from int64, to int64) {
	read := q.read.Value
	from = read & q.mask
	bufferSize = fmath.Min(bufferSize, q.size-from)
	readTo := read + bufferSize
	to = from + bufferSize
	if readTo > q.writeCache.Value {
		q.writeCache.Value = atomic.LoadInt64(&q.write.Value)
		if readTo > q.writeCache.Value {
			to = q.writeCache.Value & q.mask
		}
	}
	if from == to {
		q.failedReads.Value++
		ftime.Pause(q.pause)
	}
	q.readSize.Value = to - from
	return from, to
}

func (q *commonQ) ReleaseRead() {
	atomic.AddInt64(&q.read.Value, q.readSize.Value)
	q.readSize.Value = 0
}

func (q *commonQ) ReleaseReadLazy() {
	fatomic.LazyStore(&q.read.Value, q.read.Value+q.readSize.Value)
	q.readSize.Value = 0
}

func (q *commonQ) writeWrappingBuffer(bufferSize int64) (from int64, to int64, wrap int64) {
	writeTo := q.write.Value + bufferSize
	readLimit := writeTo - q.size
	if readLimit > q.readCache.Value {
		q.readCache.Value = atomic.LoadInt64(&q.read.Value)
		if readLimit > q.readCache.Value {
			q.failedWrites.Value++
			ftime.Pause(q.pause)
			return 0, 0, 0
		}
	}
	from = q.write.Value & q.mask
	to = fmath.Min(from+bufferSize, q.size)
	wrap = bufferSize - (to - from)
	return from, to, wrap
}

func (q *commonQ) readWrappingBuffer(bufferSize int64) (from int64, to int64, wrap int64) {
	readTo := q.read.Value + bufferSize
	if readTo > q.writeCache.Value {
		q.writeCache.Value = atomic.LoadInt64(&q.write.Value)
		if readTo > q.writeCache.Value {
			q.failedReads.Value++
			ftime.Pause(q.pause)
			return 0, 0, 0
		}
	}
	from = q.read.Value & q.mask
	to = fmath.Min(from+bufferSize, q.size)
	wrap = bufferSize - (to - from)
	return from, to, wrap
}

func (q *commonQ) FailedWrites() int64 {
	return atomic.LoadInt64(&q.failedWrites.Value)
}

func (q *commonQ) FailedReads() int64 {
	return atomic.LoadInt64(&q.failedReads.Value)
}

func (q *commonQ) String() string {
	msg := "{Size %d, mask %d, write %d, writeSize %d, failedWrites %d, readCache %d, read %d, readSize %d, failedReads %d, writeCache %d}"
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
	return fmt.Sprintf(msg, size, mask, write, writeSize, failedWrites, readCache, read, readSize, failedReads, writeCache)
}
