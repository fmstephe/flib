package main

import (
	"github.com/fmstephe/flib/queues/spscq"
	"os"
	"runtime"
	"runtime/pprof"
	"time"
)

func bqarTest(msgCount, pause, msgSize, qSize int64, profile bool) {
	q, _ := spscq.NewByteQ(qSize, pause)
	done := make(chan bool)
	if profile {
		f, err := os.Create("prof_bqar")
		if err != nil {
			panic(err.Error())
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}
	go bqarDequeue(msgCount, msgSize, q, done)
	go bqarEnqueue(msgCount, msgSize, q, done)
	<-done
	<-done
}

func bqarEnqueue(msgCount, msgSize int64, q *spscq.ByteQ, done chan bool) {
	runtime.LockOSThread()
	for i := int64(0); i < msgCount; i++ {
		writeBuffer := q.AcquireWrite(msgSize)
		for len(writeBuffer) == 0 {
			writeBuffer = q.AcquireWrite(msgSize)
		}
		writeBuffer[0] = byte(i)
		rem := msgSize - int64(len(writeBuffer))
		q.ReleaseWrite()
		for rem > 0 {
			writeBuffer = q.AcquireWrite(rem)
			rem -= int64(len(writeBuffer))
			q.ReleaseWrite()
		}
	}
	done <- true
}

func bqarDequeue(msgCount, msgSize int64, q *spscq.ByteQ, done chan bool) {
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
		q.ReleaseRead()
		for rem > 0 {
			readBuffer = q.AcquireRead(rem)
			rem -= int64(len(readBuffer))
			q.ReleaseRead()
		}
	}
	nanos := time.Now().UnixNano() - start
	printSummary(msgCount, nanos, q.FailedWrites(), q.FailedReads(), "bqar")
	expect(sum, checksum)
	done <- true
}
