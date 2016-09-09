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

	"github.com/fmstephe/flib/fmath"
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
	var buffer []unsafe.Pointer
	for t := int64(0); t < msgCount; {
		size := fmath.Min(batchSize, msgCount-t)
		buffer = q.AcquireWrite(size)
		for buffer == nil {
			buffer = q.AcquireWrite(size)
		}
		// NB: Using copy cuts about ~40% of running time
		//	copy(buffer, ptrs[t:t+size])
		for i := int64(0); i < size; i++ {
			buffer[i] = ptrs[t+i]
		}
		q.ReleaseWriteLazy()
		t += size
	}
	println("Writer done")
	done <- true
}

func pqarlDequeue(msgCount int64, q *spscq.PointerQ, batchSize int64, checksum int64, done chan bool) {
	runtime.LockOSThread()
	start := time.Now().UnixNano()
	sum := int64(0)
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
		}
		q.ReleaseReadLazy()
	}
	nanos := time.Now().UnixNano() - start
	printSummary(msgCount, nanos, q.FailedWrites(), q.FailedReads(), "pqarl")
	expect(sum, checksum)
	done <- true
}
