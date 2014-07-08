package main

import (
	"os"
	"runtime"
	"runtime/pprof"
	"time"

	"github.com/fmstephe/flib/queues/spscq"
)

func ubqTest(msgCount, msgSize, qSize int64) {
	q := spscq.NewUnsafeByteQ(qSize)
	done := make(chan bool)
	f, err := os.Create("prof_ubq")
	if err != nil {
		panic(err.Error())
	}
	pprof.StartCPUProfile(f)
	go ubqDequeue(msgCount, msgSize, q, done)
	go ubqEnqueue(msgCount, msgSize, q, done)
	<-done
	<-done
	pprof.StopCPUProfile()
}

func ubqEnqueue(msgCount, msgSize int64, q *spscq.UnsafeByteQ, done chan bool) {
	runtime.LockOSThread()
	writeBuffer := make([]byte, msgSize)
	for i := int64(0); i < msgCount; i++ {
		writeBuffer[0] = byte(i)
		for w := false; w == false; w = q.Write(writeBuffer) {
		}
	}
	done <- true
}

func ubqDequeue(msgCount, msgSize int64, q *spscq.UnsafeByteQ, done chan bool) {
	runtime.LockOSThread()
	start := time.Now().UnixNano()
	readBuffer := make([]byte, msgSize)
	sum := int64(0)
	checksum := int64(0)
	for i := int64(0); i < msgCount; i++ {
		for r := false; r == false; r = q.Read(readBuffer) {
		}
		sum += int64(readBuffer[0])
		checksum += int64(byte(i))
	}
	nanos := time.Now().UnixNano() - start
	printTimings(msgCount, nanos, "ubq")
	expect(sum, checksum)
	done <- true
}
