// Copyright 2016 Francis Stephens. All rights reserved.
// Use of this source code is governed by a BSD
// license which can be found in LICENSE.txt

package main

import (
	"flag"
	"fmt"
	"runtime"
	"unsafe"

	"github.com/fmstephe/flib/fstrconv"
)

var (
	all = flag.Bool("all", false, "Runs all queue tests")
	// ByteQ
	bqrw      = flag.Bool("bqrw", false, "Runs ByteQ using Read/Write methods")
	bqar      = flag.Bool("bqar", false, "Runs ByteQ using Acquire/Release methods")
	bqarl     = flag.Bool("bqarl", false, "Runs ByteQ with lazy Acquire/Release methods")
	bytesSize = flag.Int64("bytesSize", 63, "The number of bytes to read/write in ByteQ tests")
	// ByteChunkQ
	bcqar     = flag.Bool("bcqar", false, "Runs ByteChunkQ using Acquire/Release methods")
	bcqarl    = flag.Bool("bcqarl", false, "Runs ByteChunkQ with lazy Acquire/Release methods")
	chunkSize = flag.Int64("chunkSize", 64, "The number of bytes to read/write in ByteChunkQ tests")
	// PointerQ
	pqrw      = flag.Bool("pqrw", false, "Runs PointerQ using Read/Write methods")
	pqar      = flag.Bool("pqar", false, "Runs PointerQ using Acquire/Release methods")
	pqarl     = flag.Bool("pqarl", false, "Runs PointerQ with lazy Acquire/Release methods")
	pqs       = flag.Bool("pqs", false, "Runs PointerQ reading and writing a pointer at a time")
	pqsl      = flag.Bool("pqsl", false, "Runs PointerQ lazily reading and writing a pointer at a time")
	batchSize = flag.Int64("batchSize", 64, "The size of the read/write batches used by PointerQ")
	// Addtional flags
	millionMsgs = flag.Int64("mm", 10, "The number of messages (in millions) to send")
	qSize       = flag.Int64("qSize", 1024*1024, "The size of the queue's ring-buffer")
	pause       = flag.Int64("pause", 10*1000, "The size of the pause when a read or write fails")
	profile     = flag.Bool("profile", false, "Activates the Go profiler, outputting into a prof_* file.")
)

func main() {
	runtime.GOMAXPROCS(4)
	flag.Parse()
	var msgCount int64 = (*millionMsgs) * 1000 * 1000
	if *bqrw || *all {
		runtime.GC()
		bqrwTest(msgCount, *pause, *bytesSize, *qSize, *profile)
	}
	if *bqar || *all {
		runtime.GC()
		bqarTest(msgCount, *pause, *bytesSize, *qSize, *profile)
	}
	if *bqarl || *all {
		runtime.GC()
		bqarlTest(msgCount, *pause, *bytesSize, *qSize, *profile)
	}
	if *bcqar || *all {
		runtime.GC()
		bcqarTest(msgCount, *pause, *chunkSize, *qSize, *profile)
	}
	if *bcqarl || *all {
		runtime.GC()
		bcqarlTest(msgCount, *pause, *chunkSize, *qSize, *profile)
	}
	if *pqrw || *all {
		runtime.GC()
		pqrwTest(msgCount, *pause, *batchSize, *qSize, *profile)
	}
	if *pqar || *all {
		runtime.GC()
		pqarTest(msgCount, *pause, *batchSize, *qSize, *profile)
	}
	if *pqarl || *all {
		runtime.GC()
		pqarlTest(msgCount, *pause, *batchSize, *qSize, *profile)
	}
	if *pqs || *all {
		runtime.GC()
		pqsTest(msgCount, *pause, *qSize, *profile)
	}
	if *pqsl || *all {
		runtime.GC()
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

func getValidPointer() uintptr {
	intVal := 0
	return uintptr(unsafe.Pointer(&intVal))
}
