package main

import (
	"os"
	"runtime"
	"runtime/pprof"
	"time"

	"github.com/fmstephe/flib/queues/spscq"
)

func bcqlTest(msgCount, msgSize, qSize int64, profile bool) {
	q := spscq.NewByteChunkQ(qSize, msgSize)
	done := make(chan bool)
	if profile {
		f, err := os.Create("prof_bcql")
		if err != nil {
			panic(err.Error())
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}
	go bcqlDequeue(msgCount, q, done)
	go bcqlEnqueue(msgCount, q, done)
	<-done
	<-done
}

func bcqlEnqueue(msgCount int64, q *spscq.ByteChunkQ, done chan bool) {
	runtime.LockOSThread()
	for i := int64(0); i < msgCount; i++ {
		writeBuffer := q.WriteBuffer()
		for writeBuffer == nil {
			writeBuffer = q.WriteBuffer()
		}
		writeBuffer[0] = byte(i)
		q.CommitWriteLazy()
	}
	done <- true
}

func bcqlDequeue(msgCount int64, q *spscq.ByteChunkQ, done chan bool) {
	runtime.LockOSThread()
	start := time.Now().UnixNano()
	sum := int64(0)
	checksum := int64(0)
	for i := int64(0); i < msgCount; i++ {
		readBuffer := q.ReadBuffer()
		for readBuffer == nil {
			readBuffer = q.ReadBuffer()
		}
		sum += int64(readBuffer[0])
		checksum += int64(byte(i))
		q.CommitReadLazy()
	}
	nanos := time.Now().UnixNano() - start
	printSummary(msgCount, nanos, q.FailedWrites(), q.FailedReads(), "bcql")
	expect(sum, checksum)
	done <- true
}
