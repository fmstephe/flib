package main

import (
	"flag"
	"fmt"
	"runtime"

	"github.com/fmstephe/flib/fstrconv"
)

var (
	all = flag.Bool("all", false, "Runs all queue tests")
	// ByteQ
	byteq     = flag.Bool("bq", false, "Runs ByteQ")
	byteqLazy = flag.Bool("bql", false, "Runs ByteQ with lazy writes")
	bytesSize = flag.Int64("bytesSize", 63, "The number of bytes to read/write in ByteQ")
	// ByteChunkQ
	bytechunkq     = flag.Bool("bcq", false, "Runs ByteChunkQ")
	bytechunkqLazy = flag.Bool("bcql", false, "Runs ByteChunkQ with lazy writes")
	chunkSize      = flag.Int64("chunkSize", 64, "The number of bytes to read/write in ByteChunkQ")
	// PointerQ
	pointerq     = flag.Bool("pq", false, "Runs PointerQ")
	pointerqLazy = flag.Bool("pql", false, "Runs PointerQ with lazy writes")
	batchSize    = flag.Int64("batchSize", 1, "The size of the read/write batches used by PointerQ")
	// Addtional flags
	millionMsgs = flag.Int64("mm", 10, "The number of messages (in millions) to send")
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
	if *byteqLazy || *all {
		runtime.GC()
		bqlTest(msgCount, *bytesSize, *qSize, *profile)
	}
	if *bytechunkq || *all {
		runtime.GC()
		bcqTest(msgCount, *chunkSize, *qSize, *profile)
	}
	if *bytechunkqLazy || *all {
		runtime.GC()
		bcqlTest(msgCount, *chunkSize, *qSize, *profile)
	}
	if *pointerq || *all {
		runtime.GC()
		pqTest(msgCount, *batchSize, *qSize, *profile)
	}
	if *pointerqLazy || *all {
		runtime.GC()
		pqlTest(msgCount, *batchSize, *qSize, *profile)
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
