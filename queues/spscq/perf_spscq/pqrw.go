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

func pqrwTest(msgCount, pause, batchSize, qSize int64, profile bool) {
	ptrs, checksum := getValidPointers(msgCount)
	q, _ := spscq.NewPointerQ(qSize, pause)
	done := make(chan bool)
	if profile {
		f, err := os.Create("prof_pqrw")
		if err != nil {
			panic(err.Error())
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}
	go pqrwDequeue(msgCount, q, batchSize, checksum, done)
	go pqrwEnqueue(msgCount, q, batchSize, ptrs, done)
	<-done
	<-done
}

func pqrwEnqueue(msgCount int64, q *spscq.PointerQ, batchSize int64, ptrs []unsafe.Pointer, done chan bool) {
	runtime.LockOSThread()
	buffer := make([]unsafe.Pointer, batchSize)
	for t := int64(0); t < msgCount; t += int64(len(buffer)) {
		if batchSize > msgCount-t {
			buffer = buffer[:msgCount-t]
		}
		// NB: It cuts ~40% or run time to use copy
		// copy(buffer, ptrs[t:t+int64(len(buffer))])
		for i := range buffer {
			buffer[i] = ptrs[t+int64(i)]
		}
		for w := false; w == false; w = q.Write(buffer) {
		}
	}
	done <- true
}

func pqrwDequeue(msgCount int64, q *spscq.PointerQ, batchSize int64, checksum int64, done chan bool) {
	runtime.LockOSThread()
	start := time.Now().UnixNano()
	sum := int64(0)
	buffer := make([]unsafe.Pointer, batchSize)
	for t := int64(0); t < msgCount; t += int64(len(buffer)) {
		if batchSize > msgCount-t {
			buffer = buffer[:msgCount-t]
		}
		for r := false; r == false; r = q.Read(buffer) {
		}
		for i := range buffer {
			sum += int64(uintptr(buffer[i]))
		}
	}
	nanos := time.Now().UnixNano() - start
	printSummary(msgCount, nanos, q.FailedWrites(), q.FailedReads(), "pqrw")
	expect(sum, checksum)
	done <- true
}
