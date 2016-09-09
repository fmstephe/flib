// Copyright 2016 Francis Stephens. All rights reserved.
// Use of this source code is governed by a BSD
// license which can be found in LICENSE.txt

package main

import (
	"os"
	"runtime"
	"runtime/pprof"
	"time"
	"unsafe"

	"github.com/fmstephe/flib/queues/spscq"
)

func pqsTest(msgCount, pause, qSize int64, profile bool) {
	ptrs, checksum := getValidPointers(msgCount)
	q, _ := spscq.NewPointerQ(qSize, pause)
	done := make(chan bool)
	if profile {
		f, err := os.Create("prof_pqs")
		if err != nil {
			panic(err.Error())
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}
	go pqsDequeue(msgCount, q, checksum, done)
	go pqsEnqueue(msgCount, q, ptrs, done)
	<-done
	<-done
}

func pqsEnqueue(msgCount int64, q *spscq.PointerQ, ptrs []unsafe.Pointer, done chan bool) {
	runtime.LockOSThread()
	t := 1
	for _, ptr := range ptrs {
		w := q.WriteSingle(ptr)
		for w == false {
			w = q.WriteSingle(ptr)
		}
		t++
	}
	done <- true
}

func pqsDequeue(msgCount int64, q *spscq.PointerQ, checksum int64, done chan bool) {
	runtime.LockOSThread()
	start := time.Now().UnixNano()
	sum := int64(0)
	var v unsafe.Pointer
	for i := int64(1); i <= msgCount; i++ {
		v = q.ReadSingle()
		for v == nil {
			v = q.ReadSingle()
		}
		sum += int64(uintptr(v))
	}
	nanos := time.Now().UnixNano() - start
	printSummary(msgCount, nanos, q.FailedWrites(), q.FailedReads(), "pqs")
	expect(sum, checksum)
	done <- true
}
