package main

import (
	"os"
	"runtime"
	"runtime/pprof"
	"time"

	"github.com/fmstephe/flib/queues/spscq"
)

func bqsTest(msgCount, msgSize, qSize int64, profile bool) {
	q := spscq.NewByteQ(qSize)
	done := make(chan bool)
	if profile {
		f, err := os.Create("prof_bqs")
		if err != nil {
			panic(err.Error())
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}
	go bqsDequeue(msgCount, msgSize, q, done)
	go bqsEnqueue(msgCount, msgSize, q, done)
	<-done
	<-done
}

func bqsEnqueue(msgCount, msgSize int64, q *spscq.ByteQ, done chan bool) {
	runtime.LockOSThread()
	buffer := make([]byte, msgSize)
	for i := int64(1); i <= msgCount; i++ {
		buffer[0] = byte(i)
		for w := false; w == false; w = q.WriteSlice(buffer) {
		}
	}
	done <- true
}

func bqsDequeue(msgCount, msgSize int64, q *spscq.ByteQ, done chan bool) {
	runtime.LockOSThread()
	start := time.Now().UnixNano()
	sum := int64(0)
	checksum := int64(0)
	buffer := make([]byte, msgSize)
	for i := int64(1); i <= msgCount; i++ {
		for r := false; r == false; r = q.ReadSlice(buffer) {
		}
		sum += int64(buffer[0])
		checksum += int64(byte(i))
	}
	nanos := time.Now().UnixNano() - start
	printSummary(msgCount, nanos, q.FailedWrites(), q.FailedReads(), "bqs")
	expect(sum, checksum)
	done <- true
}
