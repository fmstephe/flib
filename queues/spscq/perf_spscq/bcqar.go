package main

import (
	"os"
	"runtime"
	"runtime/pprof"
	"time"

	"github.com/fmstephe/flib/queues/spscq"
)

func bcqarTest(msgCount, msgSize, qSize int64, profile bool) {
	q := spscq.NewByteChunkQ(qSize, msgSize)
	done := make(chan bool)
	if profile {
		f, err := os.Create("prof_bcqar")
		if err != nil {
			panic(err.Error())
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}
	go bcqarDequeue(msgCount, q, done)
	go bcqarEnqueue(msgCount, q, done)
	<-done
	<-done
}

func bcqarEnqueue(msgCount int64, q *spscq.ByteChunkQ, done chan bool) {
	runtime.LockOSThread()
	for i := int64(0); i < msgCount; i++ {
		writeBuffer := q.AcquireWrite()
		for writeBuffer == nil {
			writeBuffer = q.AcquireWrite()
		}
		writeBuffer[0] = byte(i)
		q.ReleaseWrite()
	}
	done <- true
}

func bcqarDequeue(msgCount int64, q *spscq.ByteChunkQ, done chan bool) {
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
	printSummary(msgCount, nanos, q.FailedWrites(), q.FailedReads(), "bcqar")
	expect(sum, checksum)
	done <- true
}
