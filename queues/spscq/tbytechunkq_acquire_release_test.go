// Copyright 2016 Francis Stephens. All rights reserved.
// Use of this source code is governed by a BSD
// license which can be found in LICENSE.txt

package spscq

import (
	"fmt"
	"testing"
	"time"
)

func TestSequential_ByteChunkQ(t *testing.T) {
	for sizeExp := uint(0); sizeExp <= 41; sizeExp++ {
		for chunkExp := uint(0); chunkExp < sizeExp; chunkExp++ {
			for reads := int64(1); reads <= 7; reads += 2 {
				for writes := int64(1); reads <= 7; reads += 2 {
					size := int64(1 << sizeExp)
					chunkSize := int64(1 << chunkExp)
					testSequential_ByteChunkQ(t, size, chunkSize, reads, writes)
				}
			}
		}
	}
}

func TestConcurrent_ByteChunkQ(t *testing.T) {
	for sizeExp := uint(0); sizeExp <= 41; sizeExp++ {
		for chunkExp := uint(0); chunkExp < sizeExp; chunkExp++ {
			for readSleep := time.Duration(1); readSleep <= 7; readSleep += 2 {
				for writeSleep := time.Duration(1); readSleep <= 7; readSleep += 2 {
					size := int64(1 << sizeExp)
					chunkSize := int64(1 << chunkExp)
					testConcurrentReadWrites_ByteChunkQ(t, size, chunkSize, readSleep, writeSleep)
				}
			}
		}
	}
}

func TestHeavyReadConcurrent_ByteChunkQ(t *testing.T) {
	for i := uint(0); i <= 41; i += 4 {
		size := int64(1 << i)
		for j := uint(0); j < i; j++ {
			chunkSize := int64(1 << j)
			testConcurrentReadWrites_ByteChunkQ(t, size, chunkSize, time.Microsecond, 0)
		}
	}
}

func TestHeavyWriteConcurrent_ByteChunkQ(t *testing.T) {
	for i := uint(0); i <= 41; i += 4 {
		size := int64(1 << i)
		for j := uint(0); j < i; j++ {
			chunkSize := int64(1 << j)
			testConcurrentReadWrites_ByteChunkQ(t, size, chunkSize, 0, time.Microsecond)
		}
	}
}

func testSequential_ByteChunkQ(t *testing.T, size, chunkSize, reads, writes int64) {
	iterations := int64(512)
	cqs, err := newByteChunkCommonQ(size, 0, chunkSize)
	if err != nil {
		t.Error(err.Error())
		return
	}
	cq := &cqs
	cq.initialise()
	for i := int64(0); i < iterations; i++ {
		// write
		beforeAcquireWrite := cq.write
		wfrom, wto := cq.write.bytechunkq_acquire()
		afterAcquireWrite := cq.write
		if err := testByteChunkQAcquire(wfrom, wto, beforeAcquireWrite, afterAcquireWrite); err != nil {
			t.Error(err.Error())
			return
		}
		beforeReleaseWrite := cq.write
		cq.write.release()
		afterReleaseWrite := cq.write
		if err := testRelease(beforeReleaseWrite, afterReleaseWrite); err != nil {
			t.Error(err.Error())
			return
		}
		// read
		beforeAcquireRead := cq.read
		rfrom, rto := cq.read.bytechunkq_acquire()
		afterAcquireRead := cq.read
		if err := testByteChunkQAcquire(rfrom, rto, beforeAcquireRead, afterAcquireRead); err != nil {
			t.Error(err.Error())
			return
		}
		beforeReleaseRead := cq.read
		cq.read.release()
		afterReleaseRead := cq.read
		if err := testRelease(beforeReleaseRead, afterReleaseRead); err != nil {
			t.Error(err.Error())
			return
		}
	}
}

func testConcurrentReadWrites_ByteChunkQ(t *testing.T, size, chunkSize int64, writeSleep, readSleep time.Duration) {
	iterations := int64(512)
	cqs, err := newByteChunkCommonQ(size, 0, chunkSize)
	if err != nil {
		t.Error(err.Error())
		return
	}
	cq := &cqs
	cq.initialise()
	end := make(chan bool, 2)
	go func(cq *commonQ) {
		// write
		defer func() {
			end <- true
		}()
		for i := int64(0); i < iterations; {
			// Acquire
			before := cq.write
			wfrom, wto := cq.write.bytechunkq_acquire()
			after := cq.write
			if err := testByteChunkQAcquire(wfrom, wto, before, after); err != nil {
				t.Error(err.Error())
				end <- true
				return
			}
			// Release
			before = cq.write
			cq.write.release()
			after = cq.write
			if err := testRelease(before, after); err != nil {
				t.Error(err.Error())
				end <- true
				return
			}
			i += (wto - wfrom)
			time.Sleep(writeSleep)
		}
	}(cq)
	go func(cq *commonQ) {
		// read
		defer func() {
			end <- true
		}()
		for i := int64(0); i < iterations; {
			// Acquire
			before := cq.read
			rfrom, rto := cq.read.bytechunkq_acquire()
			after := cq.read
			if err := testByteChunkQAcquire(rfrom, rto, before, after); err != nil {
				t.Error(err.Error())
				end <- true
				return
			}
			// Release
			before = cq.read
			cq.read.release()
			after = cq.read
			if err := testRelease(before, after); err != nil {
				t.Error(err.Error())
				end <- true
				return
			}
			i += (rto - rfrom)
			time.Sleep(readSleep)
		}
	}(cq)
	<-end
	<-end
}

func testByteChunkQAcquire(from, to int64, before, after acquireReleaser) error {
	if err := testReadOnlyAcquireReleaser(from, to, before, after); err != nil {
		return err
	}
	actualBufferSize := to - from
	if actualBufferSize != 0 && actualBufferSize != after.unreleased {
		return fmt.Errorf("actual buffer size (%d) not empty and does not equal after.unreleased (%d)", actualBufferSize, after.unreleased)
	}
	if actualBufferSize == 0 && before.failed+1 != after.failed {
		return fmt.Errorf("failed not incremented. Expected %d, found %d", before.failed+1, after.failed)
	}
	acquired := after.released + actualBufferSize
	available := after.oppositeCache + after.offset
	if acquired > available {
		return fmt.Errorf("Acquired (%d) greater than available (%d) by %d.\nafter %s", acquired, available, acquired-available, after.String())
	}
	if actualBufferSize == 0 && (before.released != after.oppositeCache+before.offset) { // buffer not pushing up against opposite
		return fmt.Errorf("Actual buffer size didn't need to be 0.\nbefore %s\nafter  %s", before.String(), after.String())
	}
	return nil
}