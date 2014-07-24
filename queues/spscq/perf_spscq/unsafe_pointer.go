package main

import (
	"os"
	"runtime"
	"runtime/pprof"
	"time"
	"unsafe"

	"github.com/fmstephe/flib/queues/spscq"
)

func upqTest(msgCount, batchSize, qSize int64, profile bool) {
	q := spscq.NewUnsafePointerQ(qSize)
	done := make(chan bool)
	if profile {
		f, err := os.Create("prof_upq")
		if err != nil {
			panic(err.Error())
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}
	go upqDequeue(msgCount, q, batchSize, done)
	go upqEnqueue(msgCount, q, batchSize, done)
	<-done
	<-done
}

func upqEnqueue(msgCount int64, q *spscq.UnsafePointerQ, batchSize int64, done chan bool) {
	if batchSize > 1 {
		upqBatchEnqueue(msgCount, q, batchSize, done)
	} else {
		upqSingleEnqueue(msgCount, q, done)
	}
}

func upqBatchEnqueue(msgCount int64, q *spscq.UnsafePointerQ, batchSize int64, done chan bool) {
	runtime.LockOSThread()
	var t int64
	var buffer []unsafe.Pointer
OUTER:
	for {
		buffer = q.WriteBuffer(batchSize)
		for buffer == nil {
			buffer = q.WriteBuffer(batchSize)
		}
		for i := range buffer {
			t++
			if t > msgCount {
				q.CommitWriteBuffer()
				break OUTER
			}
			buffer[i] = unsafe.Pointer(uintptr(uint(t)))
		}
		q.CommitWriteBuffer()
		if t == msgCount {
			break
		}
	}
	done <- true
}

func upqSingleEnqueue(msgCount int64, q *spscq.UnsafePointerQ, done chan bool) {
	runtime.LockOSThread()
	t := 1
	var v unsafe.Pointer
	for i := int64(0); i < msgCount; i++ {
		v = unsafe.Pointer(uintptr(uint(t)))
		w := q.WriteSingle(v)
		for w == false {
			w = q.WriteSingle(v)
		}
		t++
	}
	done <- true
}

func upqDequeue(msgCount int64, q *spscq.UnsafePointerQ, batchSize int64, done chan bool) {
	if batchSize > 1 {
		upqBatchDequeue(msgCount, q, batchSize, done)
	} else {
		upqSingleDequeue(msgCount, q, done)
	}
}

func upqBatchDequeue(msgCount int64, q *spscq.UnsafePointerQ, batchSize int64, done chan bool) {
	runtime.LockOSThread()
	start := time.Now().UnixNano()
	var sum int64
	var checksum int64
	var t int64
	var buffer []unsafe.Pointer
OUTER:
	for {
		buffer = q.ReadBuffer(batchSize)
		for buffer == nil {
			buffer = q.ReadBuffer(batchSize)
		}
		for i := range buffer {
			t++
			if t > msgCount {
				q.CommitReadBuffer()
				break OUTER
			}
			sum += int64(uintptr(buffer[i]))
			checksum += t
		}
		q.CommitReadBuffer()
		if t == msgCount {
			break
		}
	}
	nanos := time.Now().UnixNano() - start
	printTimings(msgCount, nanos, q.WriteFails(), q.ReadFails(), "upq")
	expect(sum, checksum)
	done <- true
}

func upqSingleDequeue(msgCount int64, q *spscq.UnsafePointerQ, done chan bool) {
	runtime.LockOSThread()
	start := time.Now().UnixNano()
	sum := int64(0)
	checksum := int64(0)
	var v unsafe.Pointer
	for i := int64(0); i < msgCount; i++ {
		v = q.ReadSingle()
		for v == nil {
			v = q.ReadSingle()
		}
		sum += int64(uintptr(v))
		checksum += i + 1
	}
	nanos := time.Now().UnixNano() - start
	printTimings(msgCount, nanos, q.WriteFails(), q.ReadFails(), "upq")
	expect(sum, checksum)
	done <- true
}
