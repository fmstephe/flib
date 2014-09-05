package main

import (
	"os"
	"runtime"
	"runtime/pprof"
	"time"

	"github.com/fmstephe/flib/queues/spscq"
)

func bcqarlTest(msgCount, msgSize, qSize int64, profile bool) {
	q, _ := spscq.NewByteChunkQ(qSize, msgSize)
	done := make(chan bool)
	if profile {
		f, err := os.Create("prof_bcqarl")
		if err != nil {
			panic(err.Error())
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}
	go bcqarlDequeue(msgCount, q, done)
	go bcqarlEnqueue(msgCount, q, done)
	<-done
	<-done
}

func bcqarlEnqueue(msgCount int64, q *spscq.ByteChunkQ, done chan bool) {
	runtime.LockOSThread()
	for i := int64(0); i < msgCount; i++ {
		writeBuffer := q.AcquireWrite()
		for writeBuffer == nil {
			writeBuffer = q.AcquireWrite()
		}
		writeBuffer[0] = byte(i)
		q.ReleaseWriteLazy()
	}
	done <- true
}

func bcqarlDequeue(msgCount int64, q *spscq.ByteChunkQ, done chan bool) {
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
	printSummary(msgCount, nanos, q.FailedWrites(), q.FailedReads(), "bcqarl")
	expect(sum, checksum)
	done <- true
}
