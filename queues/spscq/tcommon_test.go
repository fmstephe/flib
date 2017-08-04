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

func copyForRead(cq *commonQ) *commonQ {
	snap := &commonQ{}
	// immutable
	snap.size = cq.size
	snap.mask = cq.mask
	// write
	snap.write.released.Value = -1
	snap.write.unreleased.Value = -1
	snap.write.failed.Value = -1
	snap.write.oppositeCache.Value = -1
	// read
	snap.read.released.Value = cq.read.released.Value
	snap.read.unreleased.Value = cq.read.unreleased.Value
	snap.read.failed.Value = cq.read.failed.Value
	snap.read.oppositeCache.Value = cq.read.oppositeCache.Value
	return snap
}

func copyForWrite(cq *commonQ) *commonQ {
	snap := &commonQ{}
	// immutable
	snap.size = cq.size
	snap.mask = cq.mask
	// write
	snap.write.released.Value = cq.write.released.Value
	snap.write.unreleased.Value = cq.write.unreleased.Value
	snap.write.failed.Value = cq.write.failed.Value
	snap.write.oppositeCache.Value = cq.write.oppositeCache.Value
	// read
	snap.read.released.Value = -1
	snap.read.unreleased.Value = -1
	snap.read.failed.Value = -1
	snap.read.oppositeCache.Value = -1
	return snap
}

func testAcquireWrite(writeBufferSize, from, to int64, before, after *commonQ) error {
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
	actualWriteSize := to - from
	if actualWriteSize != after.write.unreleased.Value {
		return fmt.Errorf("actual write size (%d) does not equal after.writeSize (%d)", actualWriteSize, after.write.unreleased.Value)
	}
	if actualWriteSize == 0 && before.write.failed.Value+1 != after.write.failed.Value {
		return fmt.Errorf("failedWrites not incremented. Expected %d, found %d", before.write.failed.Value+1, after.write.failed.Value)
	}
	if actualWriteSize > writeBufferSize {
		return fmt.Errorf("Actual write size (%d) larger than requested buffer size (%d)", actualWriteSize, writeBufferSize)
	}
	if (actualWriteSize < writeBufferSize) && // Actual write smaller than asked for
		((before.write.released.Value + actualWriteSize) != (after.write.oppositeCache.Value + qSize)) && // Actual write not pushing up against read
		((before.write.released.Value+actualWriteSize)&before.mask != 0) { // Actual write not pushing against physical end of queue
		return fmt.Errorf("Actual write size (%d) could have been bigger.\nbefore %s\nafter  %s", actualWriteSize, before.writeString(), after.writeString())
	}
	if (after.write.released.Value + actualWriteSize) > (after.write.oppositeCache.Value + qSize) {
		return fmt.Errorf("Actual write size (%d) overwrites potentially unread data.\nafter %s", actualWriteSize, after.writeString())
	}
	return nil
}

func testReleaseStoredWrite(before, after *commonQ) error {
	if after.write.unreleased.Value != 0 {
		return errors.New(fmt.Sprintf("after.writeSize was not reset to 0, %d found instead", after.write.unreleased.Value))
	}
	if after.write.released.Value != before.write.released.Value+before.write.unreleased.Value {
		return errors.New(fmt.Sprintf("write has not been advanced by the correct amount.\nbefore %s\nafter  %s", before, after))
	}
	return nil
}

func testAcquireRead(readBufferSize, from, to int64, before, after *commonQ) error {
	if before.size != after.size {
		return fmt.Errorf("before.size (%d) does not equal after.size (%d)", before.size, after.size)
	}
	qSize := before.size
	if from > to {
		return errors.New(fmt.Sprintf("from (%d) is greater than to (%d)", from, to))
	}
	if from >= qSize || from < 0 {
		return errors.New(fmt.Sprintf("from (%d) must be a valid index for an array of size %d", from, qSize))
	}
	if to > qSize || to < 0 {
		return errors.New(fmt.Sprintf("to (%d) must be a valid index for an array of size %d", to, qSize))
	}
	actualReadSize := to - from
	if after.read.unreleased.Value != actualReadSize {
		return errors.New(fmt.Sprintf("after.readSize (%d) does not equal actual read size (%d)", after.read.unreleased.Value, actualReadSize))
	}
	if actualReadSize == 0 && before.read.failed.Value+1 != after.read.failed.Value {
		return errors.New(fmt.Sprintf("failedReads not incremented. Expected %d, found %d", before.read.failed.Value+1, after.read.failed.Value))
	}
	if actualReadSize > readBufferSize {
		return errors.New(fmt.Sprintf("Actual read size (%d) larger than requested buffer size (%d)", actualReadSize, readBufferSize))
	}
	if (actualReadSize < readBufferSize) && // Actual read smaller than asked for
		((before.read.released.Value + actualReadSize) != (after.read.oppositeCache.Value)) && // Actual read not pushing up against write
		((before.read.released.Value+actualReadSize)&before.mask != 0) { // Actual read not pushing against physical end of queue
		return errors.New(fmt.Sprintf("Actual read size (%d) could have been bigger.\nbefore %s\nafter  %s", actualReadSize, before.readString(), after.readString()))
	}
	if (after.read.released.Value + actualReadSize) > after.read.oppositeCache.Value {
		return errors.New(fmt.Sprintf("Actual read size (%d) reads past write position (%d).\nafter %s", actualReadSize, after.write.released.Value, after.readString()))
	}
	return nil
}

func testReleaseStoredRead(before, after *commonQ) error {
	if after.read.unreleased.Value != 0 {
		return errors.New(fmt.Sprintf("after.readSize was not reset to 0, %d found instead", after.read.unreleased.Value))
	}
	if after.read.released.Value != before.read.released.Value+before.read.unreleased.Value {
		return errors.New(fmt.Sprintf("read has not been advanced by the correct amount.\nbefore %s\nafter   %s", before, after))
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
	for j := int64(0); j < iterations; j++ {
		// write
		beforeAcquireWrite := copyForWrite(cq)
		wfrom, wto := cq.acquireWrite(writeSize)
		afterAcquireWrite := copyForWrite(cq)
		if err := testAcquireWrite(writeSize, wfrom, wto, beforeAcquireWrite, afterAcquireWrite); err != nil {
			t.Error(err.Error())
			return
		}
		beforeReleaseWrite := copyForWrite(cq)
		cq.releaseStoredWrite()
		afterReleaseWrite := copyForWrite(cq)
		if err := testReleaseStoredWrite(beforeReleaseWrite, afterReleaseWrite); err != nil {
			t.Error(err.Error())
			return
		}
		// read
		beforeAcquireRead := copyForRead(cq)
		rfrom, rto := cq.acquireRead(readSize)
		afterAcquireRead := copyForRead(cq)
		if err := testAcquireRead(readSize, rfrom, rto, beforeAcquireRead, afterAcquireRead); err != nil {
			t.Error(err.Error())
			return
		}
		beforeReleaseRead := copyForRead(cq)
		cq.releaseStoredRead()
		afterReleaseRead := copyForRead(cq)
		if err := testReleaseStoredRead(beforeReleaseRead, afterReleaseRead); err != nil {
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
	end := make(chan bool, 2)
	go func(cq *commonQ) {
		// write
		defer func() {
			end <- true
		}()
		for i := int64(0); i < iterations*writeSize; {
			before := copyForWrite(cq)
			wfrom, wto := cq.acquireWrite(writeSize)
			after := copyForWrite(cq)
			if err := testAcquireWrite(writeSize, wfrom, wto, before, after); err != nil {
				t.Error(err.Error())
				end <- true
				return
			}
			before = copyForWrite(cq)
			cq.releaseStoredWrite()
			after = copyForWrite(cq)
			if err := testReleaseStoredWrite(before, after); err != nil {
				t.Error(err.Error())
				end <- true
				return
			}
			i += (wto - wfrom)
		}
	}(&cqs)
	go func(cq *commonQ) {
		// read
		defer func() {
			end <- true
		}()
		for i := int64(0); i < iterations*writeSize; {
			before := copyForRead(cq)
			rfrom, rto := cq.acquireRead(readSize)
			after := copyForRead(cq)
			if err := testAcquireRead(readSize, rfrom, rto, before, after); err != nil {
				t.Error(err.Error())
				end <- true
				return
			}
			before = copyForRead(cq)
			cq.releaseStoredRead()
			after = copyForRead(cq)
			if err := testReleaseStoredRead(before, after); err != nil {
				t.Error(err.Error())
				end <- true
				return
			}
			i += (rto - rfrom)
		}
	}(&cqs)
	<-end
	<-end
}
