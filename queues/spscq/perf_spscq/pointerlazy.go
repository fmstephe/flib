package main

import (
	"fmt"
	"os"
	"runtime"
	"runtime/pprof"
	"time"
	"unsafe"

	"github.com/fmstephe/flib/fmath"
	"github.com/fmstephe/flib/queues/spscq"
)

func pqlTest(msgCount, batchSize, qSize int64, profile bool) {
	q := spscq.NewPointerQ(qSize)
	done := make(chan bool)
	if profile {
		f, err := os.Create("prof_pql")
		if err != nil {
			panic(err.Error())
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}
	go pqlDequeue(msgCount, q, batchSize, done)
	go pqlEnqueue(msgCount, q, batchSize, done)
	<-done
	<-done
}

func pqlEnqueue(msgCount int64, q *spscq.PointerQ, batchSize int64, done chan bool) {
	if batchSize > 1 {
		pqlBatchEnqueue(msgCount, q, batchSize, done)
	} else {
		pqlSingleEnqueue(msgCount, q, done)
	}
}

func pqlBatchEnqueue(msgCount int64, q *spscq.PointerQ, batchSize int64, done chan bool) {
	runtime.LockOSThread()
	t := int64(1)
	var buffer []unsafe.Pointer
	for t < msgCount {
		size := fmath.Min(batchSize, msgCount-t)
		buffer = q.WriteBuffer(size)
		for buffer == nil {
			buffer = q.WriteBuffer(size)
		}
		for i := range buffer {
			t++
			buffer[i] = unsafe.Pointer(uintptr(uint(t)))
		}
		q.CommitWriteLazy()
	}
	done <- true
}

func pqlSingleEnqueue(msgCount int64, q *spscq.PointerQ, done chan bool) {
	runtime.LockOSThread()
	t := 1
	var v unsafe.Pointer
	for i := int64(0); i < msgCount; i++ {
		v = unsafe.Pointer(uintptr(uint(t)))
		w := q.WriteSingleLazy(v)
		for w == false {
			w = q.WriteSingleLazy(v)
		}
		t++
	}
	done <- true
}

func pqlDequeue(msgCount int64, q *spscq.PointerQ, batchSize int64, done chan bool) {
	if batchSize > 1 {
		pqlBatchDequeue(msgCount, q, batchSize, done)
	} else {
		pqlSingleDequeue(msgCount, q, done)
	}
}

func pqlBatchDequeue(msgCount int64, q *spscq.PointerQ, batchSize int64, done chan bool) {
	runtime.LockOSThread()
	start := time.Now().UnixNano()
	sum := int64(0)
	checksum := int64(0)
	t := int64(1)
	var buffer []unsafe.Pointer
	for t < msgCount {
		buffer = q.ReadBuffer(batchSize)
		for buffer == nil {
			buffer = q.ReadBuffer(batchSize)
		}
		for i := range buffer {
			t++
			sum += int64(uintptr(buffer[i]))
			checksum += t
		}
		q.CommitReadLazy()
	}
	nanos := time.Now().UnixNano() - start
	printSummary(msgCount, nanos, q.WriteFails(), q.ReadFails(), "pql")
	expect(sum, checksum)
	done <- true
}

func pqlSingleDequeue(msgCount int64, q *spscq.PointerQ, done chan bool) {
	runtime.LockOSThread()
	start := time.Now().UnixNano()
	sum := int64(0)
	checksum := int64(0)
	var v unsafe.Pointer
	for i := int64(1); i <= msgCount; i++ {
		v = q.ReadSingleLazy()
		for v == nil {
			v = q.ReadSingleLazy()
		}
		pv := int64(uintptr(v))
		sum += pv
		checksum += i
		if pv != i {
			print(fmt.Sprintf("Bad message. Expected %d, found %d (found-expected = %d)", pv, i, pv-i))
		}
	}
	nanos := time.Now().UnixNano() - start
	printSummary(msgCount, nanos, q.WriteFails(), q.ReadFails(), "pql")
	expect(sum, checksum)
	done <- true
}
