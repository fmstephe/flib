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

func bmqarTest(msgCount, pause, msgSize, qSize int64, profile bool) {
	q, _ := spscq.NewByteMsgQ(qSize, pause)
	done := make(chan bool)
	if profile {
		f, err := os.Create("prof_bmqar")
		if err != nil {
			panic(err.Error())
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}
	go bmqarDequeue(msgCount, q, done)
	go bmqarEnqueue(msgCount, msgSize, q, done)
	<-done
	<-done
}

func bmqarEnqueue(msgCount, msgSize int64, q *spscq.ByteMsgQ, done chan bool) {
	runtime.LockOSThread()
	for i := int64(0); i < msgCount; i++ {
		writeBuffer := q.AcquireWrite(msgSize)
		for writeBuffer == nil {
			writeBuffer = q.AcquireWrite(msgSize)
		}
		writeBuffer[0] = byte(i)
		q.ReleaseWrite()
	}
	done <- true
}

func bmqarDequeue(msgCount int64, q *spscq.ByteMsgQ, done chan bool) {
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
		q.ReleaseRead()
	}
	nanos := time.Now().UnixNano() - start
	printSummary(msgCount, nanos, q.FailedWrites(), q.FailedReads(), "bmqar")
	expect(sum, checksum)
	done <- true
}
