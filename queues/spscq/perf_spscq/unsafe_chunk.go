package main

import (
	"os"
	"runtime"
	"runtime/pprof"
	"time"

	"github.com/fmstephe/flib/queues/spscq"
)

func ubcqTest(msgCount, msgSize, qSize int64, profile bool) {
	q := spscq.NewUnsafeByteChunkQ(qSize, msgSize)
	done := make(chan bool)
	if profile {
		f, err := os.Create("prof_ubcq")
		if err != nil {
			panic(err.Error())
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}
	go ubcqDequeue(msgCount, q, done)
	go ubcqEnqueue(msgCount, q, done)
	<-done
	<-done
}

func ubcqEnqueue(msgCount int64, q *spscq.UnsafeByteChunkQ, done chan bool) {
	runtime.LockOSThread()
	for i := int64(0); i < msgCount; i++ {
		writeBuffer := q.WriteBuffer()
		for writeBuffer == nil {
			writeBuffer = q.WriteBuffer()
		}
		writeBuffer[0] = byte(i)
		q.CommitWrite()
	}
	done <- true
}

func ubcqDequeue(msgCount int64, q *spscq.UnsafeByteChunkQ, done chan bool) {
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
		q.CommitRead()
	}
	nanos := time.Now().UnixNano() - start
	printTimings(msgCount, nanos, q.WriteFails(), q.ReadFails(), "ubcq")
	expect(sum, checksum)
	done <- true
}
