package main

import (
	"flag"
	"fmt"
	"runtime"

	"github.com/fmstephe/flib/fstrconv"
)

var (
	all         = flag.Bool("all", false, "Runs all queue tests")
	byteq       = flag.Bool("bq", false, "Runs ByteQ")
	bytechunkq  = flag.Bool("bcq", false, "Runs ByteChunkQ")
	pointerq    = flag.Bool("pq", false, "Runs PointerQ")
	millionMsgs = flag.Int64("mm", 10, "The number of messages (in millions) to send")
	bytesSize   = flag.Int64("bytesSize", 63, "The number of bytes to read/write in ByteQ")
	chunkSize   = flag.Int64("chunkSize", 64, "The number of bytes to read/write in ByteChunkQ")
	batchSize   = flag.Int64("batchSize", 1, "The size of the read/write batches used by PointerQ")
	qSize       = flag.Int64("qSize", 1024*1024, "The size of the queue")
	profile     = flag.Bool("profile", false, "Activates the Go profiler, outputting into a prof_* file.")
)

func main() {
	runtime.GOMAXPROCS(4)
	flag.Parse()
	var msgCount int64 = (*millionMsgs) * 1000 * 1000
	if *byteq || *all {
		runtime.GC()
		bqTest(msgCount, *bytesSize, *qSize, *profile)
	}
	if *bytechunkq || *all {
		runtime.GC()
		bcqTest(msgCount, *chunkSize, *qSize, *profile)
	}
	if *pointerq || *all {
		runtime.GC()
		pqTest(msgCount, *batchSize, *qSize, *profile)
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
