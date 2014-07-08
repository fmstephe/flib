package main

import (
	"os"
	"runtime"
	"runtime/pprof"
	"time"
	"unsafe"

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
	var t int64
	var buffer []unsafe.Pointer
OUTER:
	for {
		buffer = q.WriteBuffer(batchSize)
		for buffer == nil {
			buffer = q.WriteBuffer(batchSize)
		}
		for i := range buffer {
			t++
			if t > msgCount {
				q.CommitWriteBuffer(int64(i))
				break OUTER
			}
			buffer[i] = unsafe.Pointer(uintptr(uint(t)))
		}
		q.CommitWriteBuffer(int64(len(buffer)))
		if t == msgCount {
			break
		}
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
	var sum int64
	var checksum int64
	var t int64
	var buffer []unsafe.Pointer
OUTER:
	for {
		buffer = q.ReadBuffer(batchSize)
		for buffer == nil {
			buffer = q.ReadBuffer(batchSize)
		}
		for i := range buffer {
			t++
			if t > msgCount {
				q.CommitReadBuffer(int64(i))
				break OUTER
			}
			sum += int64(uintptr(buffer[i]))
			checksum += t
		}
		q.CommitReadBuffer(int64(len(buffer)))
		if t == msgCount {
			break
		}
	}
	nanos := time.Now().UnixNano() - start
	printTimings(msgCount, nanos, "pq")
	expect(sum, checksum)
	done <- true
}

func pqSingleDequeue(msgCount int64, q *spscq.PointerQ, done chan bool) {
	runtime.LockOSThread()
	start := time.Now().UnixNano()
	sum := int64(0)
	checksum := int64(0)
	var v unsafe.Pointer
	for i := int64(0); i < msgCount; i++ {
		v = q.ReadSingle()
		for v == nil {
			v = q.ReadSingle()
		}
		sum += int64(uintptr(v))
		checksum += i + 1
	}
	nanos := time.Now().UnixNano() - start
	printTimings(msgCount, nanos, "pq")
	expect(sum, checksum)
	done <- true
}
