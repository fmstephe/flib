package main

import (
	"os"
	"runtime"
	"runtime/pprof"
	"time"

	"github.com/fmstephe/queues/spscq"
)

func bcqTest(msgCount, msgSize, qSize int64) {
	q := spscq.NewChunkQ(qSize, msgSize)
	done := make(chan bool)
	f, err := os.Create("prof_bcq")
	if err != nil {
		panic(err.Error())
	}
	pprof.StartCPUProfile(f)
	go bcqDequeue(msgCount, q, done)
	go bcqEnqueue(msgCount, q, done)
	<-done
	<-done
	pprof.StopCPUProfile()
}

func bcqEnqueue(msgCount int64, q *spscq.ChunkQ, done chan bool) {
	runtime.LockOSThread()
	writeBuffer := q.WriteBuffer()
	for i := int64(0); i < msgCount; i++ {
		writeBuffer[0] = byte(i)
		for w := false; w == false; w = q.Write() {
		}
	}
	done <- true
}

func bcqDequeue(msgCount int64, q *spscq.ChunkQ, done chan bool) {
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
	printTimings(msgCount, nanos, "bcq")
	expect(sum, checksum)
	done <- true
}
