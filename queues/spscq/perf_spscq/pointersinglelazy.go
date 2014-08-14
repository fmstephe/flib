package main

import (
	"fmt"
	"os"
	"runtime"
	"runtime/pprof"
	"time"
	"unsafe"

	"github.com/fmstephe/flib/queues/spscq"
)

func pqsinglelTest(msgCount, qSize int64, profile bool) {
	q := spscq.NewPointerQ(qSize)
	done := make(chan bool)
	if profile {
		f, err := os.Create("prof_pqsinglel")
		if err != nil {
			panic(err.Error())
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}
	go pqsinglelDequeue(msgCount, q, done)
	go pqsinglelEnqueue(msgCount, q, done)
	<-done
	<-done
}

func pqsinglelEnqueue(msgCount int64, q *spscq.PointerQ, done chan bool) {
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

func pqsinglelDequeue(msgCount int64, q *spscq.PointerQ, done chan bool) {
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
	printSummary(msgCount, nanos, q.FailedWrites(), q.FailedReads(), "pqsinglel")
	expect(sum, checksum)
	done <- true
}
