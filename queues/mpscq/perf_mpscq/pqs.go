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

	"github.com/fmstephe/flib/queues/mpscq"
)

func pqsTest(msgCount, pause, qSize int64, profile bool) {
	ptrs, checksum := getValidPointers(msgCount)
	q, _ := mpscq.NewPointerQ(qSize, pause)
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

func pqsEnqueue(msgCount int64, q *mpscq.PointerQ, ptrs []unsafe.Pointer, done chan bool) {
	runtime.LockOSThread()
	for _, ptr := range ptrs {
		q.WriteSingleBlocking(ptr)
	}
	done <- true
}

func pqsDequeue(msgCount int64, q *mpscq.PointerQ, checksum int64, done chan bool) {
	runtime.LockOSThread()
	start := time.Now().UnixNano()
	sum := int64(0)
	for i := int64(1); i <= msgCount; i++ {
		v := q.ReadSingleBlocking()
		sum += int64(uintptr(v))
	}
	nanos := time.Now().UnixNano() - start
	printSummary(msgCount, nanos, q.FailedWrites(), q.FailedReads(), "pqs")
	expect(sum, checksum)
	done <- true
}
