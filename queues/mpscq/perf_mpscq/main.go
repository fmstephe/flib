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
	// PointerQ
	pqs = flag.Bool("pqs", false, "Runs PointerQ reading and writing a pointer at a time")
	// Addtional flags
	millionMsgs = flag.Int64("mm", 10, "The number of messages (in millions) to send")
	qSize       = flag.Int64("qSize", 1024*1024, "The size of the queue's ring-buffer")
	pause       = flag.Int64("pause", 20*1000, "The size of the pause when a read or write fails")
	profile     = flag.Bool("profile", false, "Activates the Go profiler, outputting into a prof_* file.")
)

func main() {
	runtime.GOMAXPROCS(4)
	flag.Parse()
	msgCount := (*millionMsgs) * 1e6
	debug.SetGCPercent(-1)
	if *pqs || *all {
		pqsTest(msgCount, *pause, *qSize, *profile)
	}
	runtime.GC()
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
