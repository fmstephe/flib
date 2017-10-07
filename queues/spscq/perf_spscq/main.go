// Copyright 2016 Francis Stephens. All rights reserved.
// Use of this source code is governed by a BSD
// license which can be found in LICENSE.txt

package main

import (
	"flag"
	"fmt"
	"runtime"
	"runtime/debug"
	"unsafe"

	"github.com/fmstephe/flib/fstrconv"
)

var (
	all = flag.Bool("all", false, "Runs all queue tests")
	// ByteMsgQ
	bmqar   = flag.Bool("bmqar", false, "Runs ByteMsgQ using Acquire/Release methods")
	bmqarl  = flag.Bool("bmqarl", false, "Runs ByteMsgQ using lazy Acquire/Release methods")
	msgSize = flag.Int64("msgSize", 64, "The size of messages to read/write in ByteMsgQ tests")
	// ByteChunkQ
	bcqar     = flag.Bool("bcqar", false, "Runs ByteChunkQ using Acquire/Release methods")
	bcqarl    = flag.Bool("bcqarl", false, "Runs ByteChunkQ with lazy Acquire/Release methods")
	chunkSize = flag.Int64("chunkSize", 64, "The number of bytes to read/write in ByteChunkQ tests")
	// PointerQ
	pqar      = flag.Bool("pqar", false, "Runs PointerQ using Acquire/Release methods")
	pqarl     = flag.Bool("pqarl", false, "Runs PointerQ using lazy Acquire/Release methods")
	pqs       = flag.Bool("pqs", false, "Runs PointerQ reading and writing a pointer at a time")
	pqsl      = flag.Bool("pqsl", false, "Runs PointerQ lazily reading and writing a pointer at a time")
	batchSize = flag.Int64("batchSize", 64, "The size of the read/write batches used by PointerQ")
	// Addtional flags
	msgs     = flag.Int64("m", 100, "The number of messages to send")
	msgsMult = flag.Int64("mmul", 1e6, "The multiplier for the mm argument, defaults to 1,000,000")
	qSize    = flag.Int64("qSize", 1024*1024, "The size of the queue's ring-buffer")
	pause    = flag.Int64("pause", 20*1000, "The size of the pause when a read or write fails")
	profile  = flag.Bool("profile", false, "Activates the Go profiler, outputting into a prof_* file.")
)

func main() {
	runtime.GOMAXPROCS(4)
	flag.Parse()
	msgCount := (*msgs) * (*msgsMult)
	debug.SetGCPercent(-1)
	if *bmqar || *all {
		bmqarTest(msgCount, *pause, *msgSize, *qSize, *profile)
	}
	runtime.GC()
	if *bmqarl || *all {
		bmqarlTest(msgCount, *pause, *msgSize, *qSize, *profile)
	}
	runtime.GC()
	if *bcqar || *all {
		bcqarTest(msgCount, *pause, *chunkSize, *qSize, *profile)
	}
	runtime.GC()
	if *bcqarl || *all {
		bcqarlTest(msgCount, *pause, *chunkSize, *qSize, *profile)
	}
	runtime.GC()
	if *pqar || *all {
		pqarTest(msgCount, *pause, *batchSize, *qSize, *profile)
	}
	runtime.GC()
	if *pqarl || *all {
		pqarlTest(msgCount, *pause, *batchSize, *qSize, *profile)
	}
	runtime.GC()
	if *pqs || *all {
		pqsTest(msgCount, *pause, *qSize, *profile)
	}
	runtime.GC()
	if *pqsl || *all {
		pqslTest(msgCount, *pause, *qSize, *profile)
	}
}

func printSummary(msgs, nanos, failedWrites, failedReads int64, name string) {
	sMsgs := fstrconv.ItoaComma(msgs)
	sNanos := fstrconv.ItoaComma(nanos)
	sFailedWrites := fstrconv.ItoaComma(failedWrites)
	sFailedReads := fstrconv.ItoaComma(failedReads)
	print(fmt.Sprintf("\n%s\nMsgs       %s\nNanos      %s\nfailedWrites %s\nfailedReads  %s\n", name, sMsgs, sNanos, sFailedWrites, sFailedReads))
}

func expect(sum, checksum int64) {
	if sum != checksum {
		print(fmt.Sprintf("Sum does not match checksum. sum = %d, checksum = %d\n", sum, checksum))
	}
}

func getValidPointers(num int64) (ptrs []unsafe.Pointer, checksum int64) {
	ptrs = make([]unsafe.Pointer, num)
	for i := range ptrs {
		intVal := 0
		ptrs[i] = unsafe.Pointer(&intVal)
		checksum += int64(uintptr(ptrs[i]))
	}
	return ptrs, checksum
}
