package main

import (
	"os"
	"runtime"
	"runtime/pprof"
	"time"

	"github.com/fmstephe/flib/queues/spscq"
)

func bqlTest(msgCount, msgSize, qSize int64, profile bool) {
	q := spscq.NewByteQ(qSize)
	done := make(chan bool)
	if profile {
		f, err := os.Create("prof_bql")
		if err != nil {
			panic(err.Error())
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}
	go bqlDequeue(msgCount, msgSize, q, done)
	go bqlEnqueue(msgCount, msgSize, q, done)
	<-done
	<-done
}

func bqlEnqueue(msgCount, msgSize int64, q *spscq.ByteQ, done chan bool) {
	runtime.LockOSThread()
	for i := int64(0); i < msgCount; i++ {
		writeBuffer := q.AcquireWrite(msgSize)
		for len(writeBuffer) == 0 {
			writeBuffer = q.AcquireWrite(msgSize)
		}
		writeBuffer[0] = byte(i)
		rem := msgSize - int64(len(writeBuffer))
		q.ReleaseWriteLazy()
		for rem > 0 {
			writeBuffer = q.AcquireWrite(rem)
			rem -= int64(len(writeBuffer))
			q.ReleaseWriteLazy()
		}
	}
	done <- true
}

func bqlDequeue(msgCount, msgSize int64, q *spscq.ByteQ, done chan bool) {
	runtime.LockOSThread()
	start := time.Now().UnixNano()
	sum := int64(0)
	checksum := int64(0)
	for i := int64(0); i < msgCount; i++ {
		readBuffer := q.AcquireRead(msgSize)
		for len(readBuffer) == 0 {
			readBuffer = q.AcquireRead(msgSize)
		}
		sum += int64(readBuffer[0])
		checksum += int64(byte(i))
		rem := msgSize - int64(len(readBuffer))
		q.ReleaseReadLazy()
		for rem > 0 {
			readBuffer = q.AcquireRead(rem)
			rem -= int64(len(readBuffer))
			q.ReleaseReadLazy()
		}
	}
	nanos := time.Now().UnixNano() - start
	printSummary(msgCount, nanos, q.FailedWrites(), q.FailedReads(), "bql")
	expect(sum, checksum)
	done <- true
}
