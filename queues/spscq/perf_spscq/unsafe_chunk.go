package main

import (
	"os"
	"runtime"
	"runtime/pprof"
	"time"

	"github.com/fmstephe/flib/queues/spscq"
)

func ubcqTest(msgCount, msgSize, qSize int64) {
	q := spscq.NewUnsafeByteChunkQ(qSize, msgSize)
	done := make(chan bool)
	f, err := os.Create("prof_ubcq")
	if err != nil {
		panic(err.Error())
	}
	pprof.StartCPUProfile(f)
	go ubcqDequeue(msgCount, q, done)
	go ubcqEnqueue(msgCount, q, done)
	<-done
	<-done
	pprof.StopCPUProfile()
}

func ubcqEnqueue(msgCount int64, q *spscq.UnsafeByteChunkQ, done chan bool) {
	runtime.LockOSThread()
	writeBuffer := q.WriteBuffer()
	for i := int64(0); i < msgCount; i++ {
		writeBuffer[0] = byte(i)
		for w := false; w == false; w = q.Write() {
		}
	}
	done <- true
}

func ubcqDequeue(msgCount int64, q *spscq.UnsafeByteChunkQ, done chan bool) {
	runtime.LockOSThread()
	start := time.Now().UnixNano()
	readBuffer := q.ReadBuffer()
	sum := int64(0)
	checksum := int64(0)
	for i := int64(0); i < msgCount; i++ {
		for r := false; r == false; r = q.Read() {
		}
		sum += int64(readBuffer[0])
		checksum += int64(byte(i))
	}
	nanos := time.Now().UnixNano() - start
	printTimings(msgCount, nanos, "ubcq")
	expect(sum, checksum)
	done <- true
}
