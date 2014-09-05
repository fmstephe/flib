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

func pqsTest(msgCount, qSize int64, profile bool) {
	q, _ := spscq.NewPointerQ(qSize)
	done := make(chan bool)
	if profile {
		f, err := os.Create("prof_pqs")
		if err != nil {
			panic(err.Error())
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}
	go pqsDequeue(msgCount, q, done)
	go pqsEnqueue(msgCount, q, done)
	<-done
	<-done
}

func pqsEnqueue(msgCount int64, q *spscq.PointerQ, done chan bool) {
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

func pqsDequeue(msgCount int64, q *spscq.PointerQ, done chan bool) {
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
	printSummary(msgCount, nanos, q.FailedWrites(), q.FailedReads(), "pqs")
	expect(sum, checksum)
	done <- true
}
