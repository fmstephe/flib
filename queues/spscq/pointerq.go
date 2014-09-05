package spscq

import (
	"sync/atomic"
	"unsafe"

	"github.com/fmstephe/flib/fsync/fatomic"
	"github.com/fmstephe/flib/fsync/padded"
)

type PointerQueue interface {
	// Simple Read/Write
	Read([]unsafe.Pointer) bool
	Write([]unsafe.Pointer) bool
	// Single Read/Write
	ReadSingle() unsafe.Pointer
	WriteSingle(unsafe.Pointer) bool
	ReadSingleLazy() unsafe.Pointer
	WriteSingleLazy(unsafe.Pointer) bool
	//Acquire/Release Read
	AcquireRead(int64) []unsafe.Pointer
	ReleaseRead()
	ReleaseReadLazy()
	//Acquire/Release Write
	AcquireWrite(int64) []unsafe.Pointer
	ReleaseWrite()
	ReleaseWriteLazy()
}

func NewPointerQueue(size int64) (PointerQueue, error) {
	return NewPointerQ(size)
}

type PointerQ struct {
	_prebuffer padded.CacheBuffer
	commonQ
	_midbuffer  padded.CacheBuffer
	ringBuffer  []unsafe.Pointer
	_postbuffer padded.CacheBuffer
}

func NewPointerQ(size int64) (*PointerQ, error) {
	cq, err := newCommonQ(size)
	if err != nil {
		return nil, err
	}
	ringBuffer := padded.PointerSlice(int(size))
	return &PointerQ{ringBuffer: ringBuffer, commonQ: cq}, nil
}

func (q *PointerQ) AcquireRead(bufferSize int64) []unsafe.Pointer {
	from, to := q.acquireRead(bufferSize)
	return q.ringBuffer[from:to]
}

func (q *PointerQ) AcquireWrite(bufferSize int64) []unsafe.Pointer {
	from, to := q.acquireWrite(bufferSize)
	return q.ringBuffer[from:to]
}

func (q *PointerQ) Read(buffer []unsafe.Pointer) bool {
	bufferSize := int64(len(buffer))
	from, to, wrap := q.readWrappingBuffer(bufferSize)
	if to == 0 {
		return false
	}
	copy(buffer, q.ringBuffer[from:to])
	if wrap != 0 {
		copy(buffer[bufferSize-wrap:], q.ringBuffer[:wrap])
	}
	atomic.AddInt64(&q.read.Value, bufferSize)
	return true
}

func (q *PointerQ) Write(buffer []unsafe.Pointer) bool {
	bufferSize := int64(len(buffer))
	from, to, wrap := q.writeWrappingBuffer(bufferSize)
	if to == 0 {
		return false
	}
	copy(q.ringBuffer[from:to], buffer)
	if wrap != 0 {
		copy(q.ringBuffer[:wrap], buffer[bufferSize-wrap:])
	}
	atomic.AddInt64(&q.write.Value, bufferSize)
	return true
}

func (q *PointerQ) WriteSingle(val unsafe.Pointer) bool {
	b := q.writeSingle(val)
	if b {
		atomic.AddInt64(&q.write.Value, 1)
	}
	return b
}

func (q *PointerQ) WriteSingleLazy(val unsafe.Pointer) bool {
	b := q.writeSingle(val)
	if b {
		fatomic.LazyStore(&q.write.Value, q.write.Value+1)
	}
	return b
}

func (q *PointerQ) writeSingle(val unsafe.Pointer) bool {
	write := q.write.Value
	readLimit := write - q.size
	if readLimit == q.readCache.Value {
		q.readCache.Value = atomic.LoadInt64(&q.read.Value)
		if readLimit == q.readCache.Value {
			q.failedWrites.Value++
			return false
		}
	}
	q.ringBuffer[write&q.mask] = val
	return true
}

func (q *PointerQ) ReadSingle() unsafe.Pointer {
	val := q.readSingle()
	if val != nil {
		atomic.AddInt64(&q.read.Value, 1)
	}
	return val
}

func (q *PointerQ) ReadSingleLazy() unsafe.Pointer {
	val := q.readSingle()
	if val != nil {
		fatomic.LazyStore(&q.read.Value, q.read.Value+1)
	}
	return val
}

func (q *PointerQ) readSingle() unsafe.Pointer {
	read := q.read.Value
	if read == q.writeCache.Value {
		q.writeCache.Value = atomic.LoadInt64(&q.write.Value)
		if read == q.writeCache.Value {
			q.failedReads.Value++
			return nil
		}
	}
	val := q.ringBuffer[read&q.mask]
	return val
}
