package main

import (
	"os"
	"runtime"
	"runtime/pprof"
	"time"

	"github.com/fmstephe/flib/queues/spscq"
)

func bqrwTest(msgCount, msgSize, qSize int64, profile bool) {
	q, _ := spscq.NewByteQ(qSize)
	done := make(chan bool)
	if profile {
		f, err := os.Create("prof_bqrw")
		if err != nil {
			panic(err.Error())
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}
	go bqrwDequeue(msgCount, msgSize, q, done)
	go bqrwEnqueue(msgCount, msgSize, q, done)
	<-done
	<-done
}

func bqrwEnqueue(msgCount, msgSize int64, q *spscq.ByteQ, done chan bool) {
	runtime.LockOSThread()
	buffer := make([]byte, msgSize)
	for i := int64(1); i <= msgCount; i++ {
		buffer[0] = byte(i)
		for w := false; w == false; w = q.Write(buffer) {
		}
	}
	done <- true
}

func bqrwDequeue(msgCount, msgSize int64, q *spscq.ByteQ, done chan bool) {
	runtime.LockOSThread()
	start := time.Now().UnixNano()
	sum := int64(0)
	checksum := int64(0)
	buffer := make([]byte, msgSize)
	for i := int64(1); i <= msgCount; i++ {
		for r := false; r == false; r = q.Read(buffer) {
		}
		sum += int64(buffer[0])
		checksum += int64(byte(i))
	}
	nanos := time.Now().UnixNano() - start
	printSummary(msgCount, nanos, q.FailedWrites(), q.FailedReads(), "bqrw")
	expect(sum, checksum)
	done <- true
}
