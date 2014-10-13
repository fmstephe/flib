package main

import (
	"github.com/fmstephe/flib/fmath"
	"github.com/fmstephe/flib/queues/spscq"
	"os"
	"runtime"
	"runtime/pprof"
	"time"
	"unsafe"
)

func pqarlTest(msgCount, pause, batchSize, qSize int64, profile bool) {
	q, _ := spscq.NewPointerQ(qSize, pause)
	done := make(chan bool)
	if profile {
		f, err := os.Create("prof_pqarl")
		if err != nil {
			panic(err.Error())
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}
	go pqarlDequeue(msgCount, q, batchSize, done)
	go pqarlEnqueue(msgCount, q, batchSize, done)
	<-done
	<-done
}

func pqarlEnqueue(msgCount int64, q *spscq.PointerQ, batchSize int64, done chan bool) {
	runtime.LockOSThread()
	t := int64(1)
	var buffer []unsafe.Pointer
	for t < msgCount {
		size := fmath.Min(batchSize, msgCount-t)
		buffer = q.AcquireWrite(size)
		for buffer == nil {
			buffer = q.AcquireWrite(size)
		}
		for i := range buffer {
			t++
			buffer[i] = unsafe.Pointer(uintptr(uint(t)))
		}
		q.ReleaseWriteLazy()
	}
	done <- true
}

func pqarlDequeue(msgCount int64, q *spscq.PointerQ, batchSize int64, done chan bool) {
	runtime.LockOSThread()
	start := time.Now().UnixNano()
	sum := int64(0)
	checksum := int64(0)
	t := int64(1)
	var buffer []unsafe.Pointer
	for t < msgCount {
		buffer = q.AcquireRead(batchSize)
		for buffer == nil {
			buffer = q.AcquireRead(batchSize)
		}
		for i := range buffer {
			t++
			sum += int64(uintptr(buffer[i]))
			checksum += t
		}
		q.ReleaseReadLazy()
	}
	nanos := time.Now().UnixNano() - start
	printSummary(msgCount, nanos, q.FailedWrites(), q.FailedReads(), "pqarl")
	expect(sum, checksum)
	done <- true
}
