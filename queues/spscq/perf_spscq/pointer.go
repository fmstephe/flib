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

func pqTest(msgCount, batchSize, qSize int64, profile bool) {
	q := spscq.NewPointerQ(qSize)
	done := make(chan bool)
	if profile {
		f, err := os.Create("prof_pq")
		if err != nil {
			panic(err.Error())
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}
	go pqDequeue(msgCount, q, batchSize, done)
	go pqEnqueue(msgCount, q, batchSize, done)
	<-done
	<-done
}

func pqEnqueue(msgCount int64, q *spscq.PointerQ, batchSize int64, done chan bool) {
	if batchSize > 1 {
		pqBatchEnqueue(msgCount, q, batchSize, done)
	} else {
		pqSingleEnqueue(msgCount, q, done)
	}
}

func pqBatchEnqueue(msgCount int64, q *spscq.PointerQ, batchSize int64, done chan bool) {
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
		q.CommitWriteBuffer()
	}
	done <- true
}

func pqSingleEnqueue(msgCount int64, q *spscq.PointerQ, done chan bool) {
	runtime.LockOSThread()
	t := 1
	var v unsafe.Pointer
	for i := int64(0); i < msgCount; i++ {
		v = unsafe.Pointer(uintptr(uint(t)))
		w := q.WriteSingle(v)
		for w == false {
			w = q.WriteSingle(v)
		}
		t++
	}
	done <- true
}

func pqDequeue(msgCount int64, q *spscq.PointerQ, batchSize int64, done chan bool) {
	if batchSize > 1 {
		pqBatchDequeue(msgCount, q, batchSize, done)
	} else {
		pqSingleDequeue(msgCount, q, done)
	}
}

func pqBatchDequeue(msgCount int64, q *spscq.PointerQ, batchSize int64, done chan bool) {
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
		q.CommitReadBuffer()
	}
	nanos := time.Now().UnixNano() - start
	printTimings(msgCount, nanos, q.WriteFails(), q.ReadFails(), "pq")
	expect(sum, checksum)
	done <- true
}

func pqSingleDequeue(msgCount int64, q *spscq.PointerQ, done chan bool) {
	runtime.LockOSThread()
	start := time.Now().UnixNano()
	sum := int64(0)
	checksum := int64(0)
	var v unsafe.Pointer
	for i := int64(1); i <= msgCount; i++ {
		v = q.ReadSingle()
		for v == nil {
			v = q.ReadSingle()
		}
		pv := int64(uintptr(v))
		sum += pv
		checksum += i
		if pv != i {
			print(fmt.Sprintf("Bad message. Expected %d, found %d (found-expected = %d)", pv, i, pv-i))
		}
	}
	nanos := time.Now().UnixNano() - start
	printTimings(msgCount, nanos, q.WriteFails(), q.ReadFails(), "pq")
	expect(sum, checksum)
	done <- true
}
