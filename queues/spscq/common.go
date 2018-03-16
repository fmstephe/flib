// Copyright 2016 Francis Stephens. All rights reserved.
// Use of this source code is governed by a BSD
// license which can be found in LICENSE.txt

package spscq

import (
	"errors"
	"fmt"

	"github.com/fmstephe/flib/fmath"
)

const maxSize = 1 << 41

type commonQ struct {
	// Readonly Fields
	// TODO Remove these
	size  int64
	mask  int64
	pause int64
	// Writer fields
	write acquireReleaser
	// Reader fields
	read acquireReleaser
}

func newPointerCommonQ(size, pause int64) (commonQ, error) {
	return newCommonQ(size, pause)
}

func newByteChunkCommonQ(size, pause, chunkSize int64) (commonQ, error) {
	cq, err := newCommonQ(size, pause)
	cq.write.chunk = chunkSize
	cq.read.chunk = chunkSize
	return cq, err
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
		write: acquireReleaser{name: "write", size: size, mask: size - 1, pause: pause},
		read:  acquireReleaser{name: "read", size: size, mask: size - 1, pause: pause},
	}
	q.write.offset = size
	return q, nil
}

func (q *commonQ) initialise() {
	q.write.opposite = &q.read.released
	q.read.opposite = &q.write.released
}

func (q *commonQ) String() string {
	return fmt.Sprintf("%s, %s", &q.read, &q.write)
}

func (q *commonQ) FailedWrites() int64 {
	return q.write.getFailed()
}

func (q *commonQ) FailedReads() int64 {
	return q.read.getFailed()
}
