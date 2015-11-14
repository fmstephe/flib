package spscq

import (
	"runtime/debug"
	"testing"
	"unsafe"
)

func BenchmarkStrict(b *testing.B) {
	println(b.N)
	debug.SetGCPercent(-1)
	q, _ := NewPointerQ(64*1024, 10*1000)
	done := make(chan int64)
	b.ResetTimer()
	go pqsDequeue(int64(b.N), q, done)
	go pqsEnqueue(int64(b.N), q, done)
	<-done
	<-done
	b.StopTimer()
	// Clear out all those rediculous pointers before we garbage collect
	for i := range q.ringBuffer {
		q.ringBuffer[i] = unsafe.Pointer(uintptr(0))
	}
	debug.SetGCPercent(100)
}

func BenchmarkLazy(b *testing.B) {
	println(b.N)
	debug.SetGCPercent(-1)
	q, _ := NewPointerQ(64*1024, 10*1000)
	done := make(chan int64)
	b.ResetTimer()
	go pqsDequeueLazy(int64(b.N), q, done)
	go pqsEnqueueLazy(int64(b.N), q, done)
	<-done
	<-done
	b.StopTimer()
	// Clear out all those rediculous pointers before we garbage collect
	for i := range q.ringBuffer {
		q.ringBuffer[i] = unsafe.Pointer(uintptr(0))
	}
	debug.SetGCPercent(100)
}

func pqsEnqueue(msgCount int64, q *PointerQ, done chan int64) {
	t := 1
	var v unsafe.Pointer
	for i := int64(0); i < msgCount; i++ {
		v = unsafe.Pointer(uintptr(uint64(t)))
		w := q.WriteSingle(v)
		for w == false {
			w = q.WriteSingle(v)
		}
		t++
	}
	done <- -1
}

func pqsDequeue(msgCount int64, q *PointerQ, done chan int64) {
	sum := int64(0)
	var v unsafe.Pointer
	for i := int64(1); i <= msgCount; i++ {
		v = q.ReadSingle()
		for v == nil {
			v = q.ReadSingle()
		}
		pv := int64(uintptr(v))
		sum += pv
	}
	done <- sum
}

func pqsEnqueueLazy(msgCount int64, q *PointerQ, done chan int64) {
	t := 1
	var v unsafe.Pointer
	for i := int64(0); i < msgCount; i++ {
		v = unsafe.Pointer(uintptr(uint64(t)))
		w := q.WriteSingleLazy(v)
		for w == false {
			w = q.WriteSingleLazy(v)
		}
		t++
	}
	done <- -1
}

func pqsDequeueLazy(msgCount int64, q *PointerQ, done chan int64) {
	sum := int64(0)
	var v unsafe.Pointer
	for i := int64(1); i <= msgCount; i++ {
		v = q.ReadSingleLazy()
		for v == nil {
			v = q.ReadSingleLazy()
		}
		pv := int64(uintptr(v))
		sum += pv
	}
	done <- sum
}
