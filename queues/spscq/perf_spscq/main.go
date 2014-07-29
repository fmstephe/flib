package main

import (
	"flag"
	"fmt"
	"runtime"

	"github.com/fmstephe/flib/fstrconv"
)

var (
	all         = flag.Bool("all", false, "Runs all queue tests")
	bytechunkq  = flag.Bool("bcq", false, "Runs ByteChunkQ")
	pointerq    = flag.Bool("pq", false, "Runs PointerQ")
	ubytechunkq = flag.Bool("ubcq", false, "Runs UnsafeByteChunkQ")
	upointerq   = flag.Bool("upq", false, "Runs UnsafePointerQ")
	millionMsgs = flag.Int64("mm", 10, "The number of messages (in millions) to send")
	chunkSize   = flag.Int64("chunkSize", 64, "The number of bytes to read/write in ByteChunkQ")
	batchSize   = flag.Int64("batchSize", 1, "The size of the read/write batches used by PointerQ")
	qSize       = flag.Int64("qSize", 1024*1024, "The size of the queue")
	profile     = flag.Bool("profile", false, "Activates the Go profiler, outputting into a prof_* file.")
)

func main() {
	runtime.GOMAXPROCS(4)
	flag.Parse()
	var msgCount int64 = (*millionMsgs) * 1000 * 1000
	if *bytechunkq || *all {
		runtime.GC()
		bcqTest(msgCount, *chunkSize, *qSize, *profile)
	}
	if *ubytechunkq || *all {
		runtime.GC()
		ubcqTest(msgCount, *chunkSize, *qSize, *profile)
	}
	if *pointerq || *all {
		runtime.GC()
		pqTest(msgCount, *batchSize, *qSize, *profile)
	}
	if *upointerq || *all {
		runtime.GC()
		upqTest(msgCount, *batchSize, *qSize, *profile)
	}
}

func printTimings(msgs, nanos, writeFails, readFails int64, name string) {
	sMsgs := fstrconv.ItoaComma(msgs)
	sNanos := fstrconv.ItoaComma(nanos)
	sWriteFails := fstrconv.ItoaComma(writeFails)
	sReadFails := fstrconv.ItoaComma(readFails)
	print(fmt.Sprintf("\n%s\nMsgs       %s\nNanos      %s\nwriteFails %s\nreadFails  %s\n", name, sMsgs, sNanos, sWriteFails, sReadFails))
}

func expect(sum, checksum int64) {
	if sum != checksum {
		print(fmt.Sprintf("Sum does not match checksum. sum = %d, checksum = %d\n", sum, checksum))
	}
}
