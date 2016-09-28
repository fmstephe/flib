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

func pqarlTest(msgCount, pause, batchSize, qSize int64, profile bool) {
	ptrs, checksum := getValidPointers(msgCount)
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
	go pqarlDequeue(msgCount, q, batchSize, checksum, done)
	go pqarlEnqueue(msgCount, q, batchSize, ptrs, done)
	<-done
	<-done
}

func pqarlEnqueue(msgCount int64, q *spscq.PointerQ, batchSize int64, ptrs []unsafe.Pointer, done chan bool) {
	runtime.LockOSThread()
	for t := int64(0); t < msgCount; {
		if batchSize > msgCount-t {
			batchSize = msgCount - t
		}
		buffer := q.AcquireWrite(batchSize)
		// NB: It cuts ~40% of run time to use copy
		copy(buffer, ptrs[t:t+int64(len(buffer))])
		/*
			for i := range buffer {
				buffer[i] = ptrs[t+int64(i)]
			}
		*/
		q.ReleaseWriteLazy()
		t += int64(len(buffer))
	}
	done <- true
}

func pqarlDequeue(msgCount int64, q *spscq.PointerQ, batchSize int64, checksum int64, done chan bool) {
	runtime.LockOSThread()
	start := time.Now().UnixNano()
	sum := int64(0)
	for t := int64(0); t < msgCount; {
		buffer := q.AcquireRead(batchSize)
		for i := range buffer {
			sum += int64(uintptr(buffer[i]))
		}
		q.ReleaseReadLazy()
		t += int64(len(buffer))
	}
	nanos := time.Now().UnixNano() - start
	printSummary(msgCount, nanos, q.FailedWrites(), q.FailedReads(), "pqarl")
	expect(sum, checksum)
	done <- true
}
