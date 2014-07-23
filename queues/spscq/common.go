package spscq

import (
	"github.com/fmstephe/flib/fsync/padded"
	"sync/atomic"
)

type paddedCounters struct {
	_prebuffer padded.CacheBuffer
	// Writer fields
	write     padded.Int64
	writeFail padded.Int64
	readCache padded.Int64
	// Reader fields
	read       padded.Int64
	readFail   padded.Int64
	writeCache padded.Int64
	_midbuffer padded.CacheBuffer
}

func (c *paddedCounters) WriteFails() int64 {
	return atomic.LoadInt64(&c.writeFail.Value)
}

func (c *paddedCounters) ReadFails() int64 {
	return atomic.LoadInt64(&c.readFail.Value)
}
