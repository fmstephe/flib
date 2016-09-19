// Copyright 2016 Francis Stephens. All rights reserved.
// Use of this source code is governed by a BSD
// license which can be found in LICENSE.txt

package main

import (
	"os"
	"runtime"
	"runtime/pprof"
	"time"

	"github.com/fmstephe/flib/queues/spscq"
)

func bmqarlTest(msgCount, pause, msgSize, qSize int64, profile bool) {
	q, _ := spscq.NewByteMsgQ(qSize, pause)
	done := make(chan bool)
	if profile {
		f, err := os.Create("prof_bmqarl")
		if err != nil {
			panic(err.Error())
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}
	go bmqarlDequeue(msgCount, q, done)
	go bmqarlEnqueue(msgCount, msgSize, q, done)
	<-done
	<-done
}

func bmqarlEnqueue(msgCount, msgSize int64, q *spscq.ByteMsgQ, done chan bool) {
	runtime.LockOSThread()
	for i := int64(0); i < msgCount; i++ {
		writeBuffer := q.AcquireWrite(msgSize)
		for writeBuffer == nil {
			writeBuffer = q.AcquireWrite(msgSize)
		}
		writeBuffer[0] = byte(i)
		q.ReleaseWriteLazy()
	}
	done <- true
}

func bmqarlDequeue(msgCount int64, q *spscq.ByteMsgQ, done chan bool) {
	runtime.LockOSThread()
	start := time.Now().UnixNano()
	sum := int64(0)
	checksum := int64(0)
	for i := int64(0); i < msgCount; i++ {
		readBuffer := q.AcquireRead()
		for readBuffer == nil {
			readBuffer = q.AcquireRead()
		}
		sum += int64(readBuffer[0])
		checksum += int64(byte(i))
		q.ReleaseReadLazy()
	}
	nanos := time.Now().UnixNano() - start
	printSummary(msgCount, nanos, q.FailedWrites(), q.FailedReads(), "bmqarl")
	expect(sum, checksum)
	done <- true
}
