package spscq

import (
	"fmt"
	"github.com/fmstephe/flib/fmath"
	"github.com/fmstephe/flib/fsync/padded"
	"sync/atomic"
)

type commonQ struct {
	// Readonly Fields
	size       int64
	mask       int64
	_ropadding padded.CacheBuffer
	// Writer fields
	write     padded.Int64
	writeSize padded.Int64
	writeFail padded.Int64
	readCache padded.Int64
	// Reader fields
	read       padded.Int64
	readSize   padded.Int64
	readFail   padded.Int64
	writeCache padded.Int64
}

func newCommonQ(size int64) commonQ {
	if !fmath.PowerOfTwo(size) {
		panic(fmt.Sprintf("Size (%d) must be a power of two", size))
	}
	cq := commonQ{size: size, mask: size - 1}
	return cq
}

func (c *commonQ) WriteFails() int64 {
	return atomic.LoadInt64(&c.writeFail.Value)
}

func (c *commonQ) ReadFails() int64 {
	return atomic.LoadInt64(&c.readFail.Value)
}
