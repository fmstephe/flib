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
		writeBuffer := q.WriteBuffer(msgSize)
		for len(writeBuffer) == 0 {
			writeBuffer = q.WriteBuffer(msgSize)
		}
		writeBuffer[0] = byte(i)
		rem := msgSize - int64(len(writeBuffer))
		q.CommitWriteLazy()
		for rem > 0 {
			writeBuffer = q.WriteBuffer(rem)
			rem -= int64(len(writeBuffer))
			q.CommitWriteLazy()
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
		readBuffer := q.ReadBuffer(msgSize)
		for len(readBuffer) == 0 {
			readBuffer = q.ReadBuffer(msgSize)
		}
		sum += int64(readBuffer[0])
		checksum += int64(byte(i))
		rem := msgSize - int64(len(readBuffer))
		q.CommitReadLazy()
		for rem > 0 {
			readBuffer = q.ReadBuffer(rem)
			rem -= int64(len(readBuffer))
			q.CommitReadLazy()
		}
	}
	nanos := time.Now().UnixNano() - start
	printTimings(msgCount, nanos, q.ReadFails(), q.ReadFails(), "bql")
	expect(sum, checksum)
	done <- true
}
