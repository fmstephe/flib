// Copyright 2016 Francis Stephens. All rights reserved.
// Use of this source code is governed by a BSD
// license which can be found in LICENSE.txt

package spscq

import (
	"fmt"
	"testing"
)

func TestSequential_PointerQ(t *testing.T) {
	for sizeExp := uint(0); sizeExp <= 41; sizeExp++ {
		for readsExp := uint(1); readsExp <= 7; readsExp += 2 {
			for writesExp := uint(1); readsExp <= 7; readsExp += 2 {
				size := int64(1 << sizeExp)
				readSize := int64(1 << readsExp)
				writeSize := int64(1 << writesExp)
				testSequential_PointerQ(t, size, readSize, writeSize)
			}
		}
	}
}

func TestConcurrent_PointerQ(t *testing.T) {
	for sizeExp := uint(0); sizeExp <= 41; sizeExp++ {
		for readsExp := uint(1); readsExp <= 7; readsExp += 2 {
			for writesExp := uint(1); readsExp <= 7; readsExp += 2 {
				size := int64(1 << sizeExp)
				readSize := int64(1 << readsExp)
				writeSize := int64(1 << writesExp)
				testConcurrent_PointerQ(t, size, readSize, writeSize)
			}
		}
	}
}

func testSequential_PointerQ(t *testing.T, size, readSize, writeSize int64) {
	cqs, err := newPointerCommonQ(size, 0)
	if err != nil {
		t.Error(err.Error())
		return
	}
	cq := &cqs
	cq.initialise()
	for j := int64(0); j < 512; j++ {
		// write
		beforeAcquireWrite := cq.write
		wfrom, wto := cq.write.pointerq_acquire(writeSize)
		afterAcquireWrite := cq.write
		if err := testPointerQAcquire(writeSize, wfrom, wto, beforeAcquireWrite, afterAcquireWrite); err != nil {
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
		rfrom, rto := cq.read.pointerq_acquire(readSize)
		afterAcquireRead := cq.read
		if err := testPointerQAcquire(readSize, rfrom, rto, beforeAcquireRead, afterAcquireRead); err != nil {
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

func testConcurrent_PointerQ(t *testing.T, size, readSize, writeSize int64) {
	iterations := int64(512)
	cqs, err := newPointerCommonQ(size, 0)
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
		for i := int64(0); i < iterations*writeSize; {
			before := cq.write
			wfrom, wto := cq.write.pointerq_acquire(writeSize)
			after := cq.write
			if err := testPointerQAcquire(writeSize, wfrom, wto, before, after); err != nil {
				t.Error(err.Error())
				end <- true
				return
			}
			before = cq.write
			cq.write.release()
			after = cq.write
			if err := testRelease(before, after); err != nil {
				t.Error(err.Error())
				end <- true
				return
			}
			i += (wto - wfrom)
		}
	}(cq)
	go func(cq *commonQ) {
		// read
		defer func() {
			end <- true
		}()
		for i := int64(0); i < iterations*writeSize; {
			before := cq.read
			rfrom, rto := cq.read.pointerq_acquire(readSize)
			after := cq.read
			if err := testPointerQAcquire(readSize, rfrom, rto, before, after); err != nil {
				t.Error(err.Error())
				end <- true
				return
			}
			before = cq.read
			cq.read.release()
			after = cq.read
			if err := testRelease(before, after); err != nil {
				t.Error(err.Error())
				end <- true
				return
			}
			i += (rto - rfrom)
		}
	}(cq)
	<-end
	<-end
}

func testPointerQAcquire(requestedBufferSize, from, to int64, before, after acquireReleaser) error {
	if err := testReadOnlyAcquireReleaser(from, to, before, after); err != nil {
		return err
	}
	actualBufferSize := to - from
	if actualBufferSize != after.unreleased {
		return fmt.Errorf("actual buffer size (%d) does not equal after.unreleased (%d)", actualBufferSize, after.unreleased)
	}
	if actualBufferSize == 0 && before.failed+1 != after.failed {
		return fmt.Errorf("failed not incremented. Expected %d, found %d", before.failed+1, after.failed)
	}
	if actualBufferSize > requestedBufferSize {
		return fmt.Errorf("Actual buffer size (%d) larger than requested buffer size (%d)", actualBufferSize, requestedBufferSize)
	}
	if (actualBufferSize < requestedBufferSize) && // buffer smaller than asked for
		((before.released + actualBufferSize) != (after.oppositeCache + before.offset)) && // buffer not pushing up against opposite
		((before.released+actualBufferSize)&before.mask != 0) { // buffer not pushing against physical end of queue
		return fmt.Errorf("Actual buffer size (%d) could have been bigger.\nbefore %s\nafter  %s", actualBufferSize, before.String(), after.String())
	}
	if (after.released + actualBufferSize) > (after.oppositeCache + before.size) {
		return fmt.Errorf("Actual buffer size (%d) overwrites potentially unread data.\nafter %s", actualBufferSize, after.String())
	}
	return nil
}
