// Copyright 2016 Francis Stephens. All rights reserved.
// Use of this source code is governed by a BSD
// license which can be found in LICENSE.txt

package spscq

import (
	"errors"
	"fmt"
	"testing"

	"github.com/fmstephe/flib/fmath"
)

// Test that we can call newCommonQ(...) for every power of 2 in an int64
func TestNewCommonQPowerOf2(t *testing.T) {
	for size := int64(1); size <= maxSize; size *= 2 {
		_, err := newCommonQ(size, 0)
		if err != nil {
			t.Errorf("Error found for size %d", size)
		}
	}
}

// Test that we can't call newCommonQ(...) with a non-power of 2 size
func TestNewCommonQNotPowerOf2(t *testing.T) {
	for size := int64(1); size < 10*1000; size++ {
		if !fmath.PowerOfTwo(size) {
			_, err := newCommonQ(size, 0)
			if err == nil {
				t.Errorf("No error detected for size %d", size)
			}
		}
	}
}

func testAcquire(requestedBufferSize, from, to int64, before, after mutableFields) error {
	if before.size != after.size {
		return fmt.Errorf("before.size (%d) does not equal after.size (%d)", before.size, after.size)
	}
	qSize := before.size
	if from >= qSize || from < 0 {
		return fmt.Errorf("from (%d) must be a valid index for an array of size %d", from, qSize)
	}
	if to > qSize || to < 0 {
		return fmt.Errorf("to (%d) must be a valid index for an array of size %d", to, qSize)
	}
	if from > to {
		return fmt.Errorf("from (%d) is greater than to (%d)", from, to)
	}
	actualBufferSize := to - from
	if actualBufferSize != after.unreleased.Value {
		return fmt.Errorf("actual write size (%d) does not equal after.writeSize (%d)", actualBufferSize, after.unreleased.Value)
	}
	if actualBufferSize == 0 && before.failed.Value+1 != after.failed.Value {
		return fmt.Errorf("failedWrites not incremented. Expected %d, found %d", before.failed.Value+1, after.failed.Value)
	}
	if actualBufferSize > requestedBufferSize {
		return fmt.Errorf("Actual write size (%d) larger than requested buffer size (%d)", actualBufferSize, requestedBufferSize)
	}
	if (actualBufferSize < requestedBufferSize) && // buffer smaller than asked for
		((before.released.Value + actualBufferSize) != (after.oppositeCache.Value + before.offset.Value)) && // buffer not pushing up against opposite
		((before.released.Value+actualBufferSize)&before.mask != 0) { // buffer not pushing against physical end of queue
		return fmt.Errorf("Actual write size (%d) could have been bigger.\nbefore %s\nafter  %s", actualBufferSize, before.String(), after.String())
	}
	if (after.released.Value + actualBufferSize) > (after.oppositeCache.Value + qSize) {
		return fmt.Errorf("Actual write size (%d) overwrites potentially unread data.\nafter %s", actualBufferSize, after.String())
	}
	return nil
}

func testRelease(before, after mutableFields) error {
	if after.unreleased.Value != 0 {
		return errors.New(fmt.Sprintf("unreleased was not reset to 0, %d found instead", after.unreleased.Value))
	}
	if after.released.Value != before.released.Value+before.unreleased.Value {
		return errors.New(fmt.Sprintf("released has not been advanced by the correct amount.\nbefore %s\nafter  %s", before, after))
	}
	return nil
}

func TestEvenWriteRead(t *testing.T) {
	for i := uint(0); i <= 41; i += 4 {
		size := int64(1 << i)
		bufferSize := fmath.Max(size/128, 1)
		testSequentialReadWrites(t, size, bufferSize, bufferSize, 512)
	}
}

func TestLightWriteHeavyRead(t *testing.T) {
	for i := uint(0); i <= 41; i += 4 {
		size := int64(1 << i)
		bufferSize := fmath.Max(size/128, 1)
		testSequentialReadWrites(t, size, bufferSize, bufferSize*2, 512)
	}
}

func TestHeavyWriteLightRead(t *testing.T) {
	for i := uint(0); i <= 41; i += 4 {
		size := int64(1 << i)
		bufferSize := fmath.Max(size/128, 1)
		testSequentialReadWrites(t, size, bufferSize*2, bufferSize, 512)
	}
}

func testSequentialReadWrites(t *testing.T, size int64, writeSize, readSize, iterations int64) {
	cqs, err := newCommonQ(size, 0)
	if err != nil {
		t.Error(err.Error())
		return
	}
	cq := &cqs
	cq.initialise()
	for j := int64(0); j < iterations; j++ {
		// write
		beforeAcquireWrite := cq.write
		wfrom, wto := cq.acquireWrite(writeSize)
		afterAcquireWrite := cq.write
		if err := testAcquire(writeSize, wfrom, wto, beforeAcquireWrite, afterAcquireWrite); err != nil {
			t.Error(err.Error())
			return
		}
		beforeReleaseWrite := cq.write
		cq.releaseStoredWrite()
		afterReleaseWrite := cq.write
		if err := testRelease(beforeReleaseWrite, afterReleaseWrite); err != nil {
			t.Error(err.Error())
			return
		}
		// read
		beforeAcquireRead := cq.read
		rfrom, rto := cq.acquireRead(readSize)
		afterAcquireRead := cq.read
		if err := testAcquire(readSize, rfrom, rto, beforeAcquireRead, afterAcquireRead); err != nil {
			t.Error(err.Error())
			return
		}
		beforeReleaseRead := cq.read
		cq.releaseStoredRead()
		afterReleaseRead := cq.read
		if err := testRelease(beforeReleaseRead, afterReleaseRead); err != nil {
			t.Error(err.Error())
			return
		}
	}
}

func TestEvenWriteReadConc(t *testing.T) {
	for i := uint(0); i <= 41; i += 4 {
		size := int64(1 << i)
		bufferSize := fmath.Max(size/128, 1)
		testConcurrentReadWrites(t, size, bufferSize, bufferSize, 512)
	}
}

func TestLightWriteHeavyReadConc(t *testing.T) {
	for i := uint(0); i <= 41; i += 4 {
		size := int64(1 << i)
		bufferSize := fmath.Max(size/128, 1)
		testConcurrentReadWrites(t, size, bufferSize, bufferSize*2, 512)
	}
}

func TestHeavyWriteLightReadConc(t *testing.T) {
	for i := uint(0); i <= 41; i += 4 {
		size := int64(1 << i)
		bufferSize := fmath.Max(size/128, 1)
		testConcurrentReadWrites(t, size, bufferSize*2, bufferSize, 512)
	}
}

func testConcurrentReadWrites(t *testing.T, size int64, writeSize, readSize, iterations int64) {
	cqs, err := newCommonQ(size, 0)
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
			wfrom, wto := cq.acquireWrite(writeSize)
			after := cq.write
			if err := testAcquire(writeSize, wfrom, wto, before, after); err != nil {
				t.Error(err.Error())
				end <- true
				return
			}
			before = cq.write
			cq.releaseStoredWrite()
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
			rfrom, rto := cq.acquireRead(readSize)
			after := cq.read
			if err := testAcquire(readSize, rfrom, rto, before, after); err != nil {
				t.Error(err.Error())
				end <- true
				return
			}
			before = cq.read
			cq.releaseStoredRead()
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
