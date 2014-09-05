package main

import (
	"os"
	"runtime"
	"runtime/pprof"
	"time"
	"unsafe"

	"github.com/fmstephe/flib/queues/spscq"
)

func pqrwTest(msgCount, batchSize, qSize int64, profile bool) {
	q, _ := spscq.NewPointerQ(qSize)
	done := make(chan bool)
	if profile {
		f, err := os.Create("prof_pqrw")
		if err != nil {
			panic(err.Error())
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}
	go pqrwDequeue(msgCount, q, batchSize, done)
	go pqrwEnqueue(msgCount, q, batchSize, done)
	<-done
	<-done
}

func pqrwEnqueue(msgCount int64, q *spscq.PointerQ, batchSize int64, done chan bool) {
	runtime.LockOSThread()
	t := int64(1)
	buffer := make([]unsafe.Pointer, batchSize)
	for t < msgCount {
		if batchSize > msgCount-t {
			buffer = buffer[:msgCount-t]
		}
		for i := range buffer {
			t++
			buffer[i] = unsafe.Pointer(uintptr(uint(t)))
		}
		for w := false; w == false; w = q.Write(buffer) {
		}
	}
	done <- true
}

func pqrwDequeue(msgCount int64, q *spscq.PointerQ, batchSize int64, done chan bool) {
	runtime.LockOSThread()
	start := time.Now().UnixNano()
	sum := int64(0)
	checksum := int64(0)
	t := int64(1)
	buffer := make([]unsafe.Pointer, batchSize)
	for t < msgCount {
		if batchSize > msgCount-t {
			buffer = buffer[:msgCount-t]
		}
		for r := false; r == false; r = q.Read(buffer) {
		}
		for i := range buffer {
			t++
			sum += int64(uintptr(buffer[i]))
			checksum += t
		}
	}
	nanos := time.Now().UnixNano() - start
	printSummary(msgCount, nanos, q.FailedWrites(), q.FailedReads(), "pqrw")
	expect(sum, checksum)
	done <- true
}
