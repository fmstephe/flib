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
	snap.write.Value = -1
	snap.writeSize.Value = -1
	snap.failedWrites.Value = -1
	snap.readCache.Value = -1
	// read
	snap.read.Value = cq.read.Value
	snap.readSize.Value = cq.readSize.Value
	snap.failedReads.Value = cq.failedReads.Value
	snap.writeCache.Value = cq.writeCache.Value
	return snap
}

func copyForWrite(cq *commonQ) *commonQ {
	snap := &commonQ{}
	// immutable
	snap.size = cq.size
	snap.mask = cq.mask
	// write
	snap.write.Value = cq.write.Value
	snap.writeSize.Value = cq.writeSize.Value
	snap.failedWrites.Value = cq.failedWrites.Value
	snap.readCache.Value = cq.readCache.Value
	// read
	snap.read.Value = -1
	snap.readSize.Value = -1
	snap.failedReads.Value = -1
	snap.writeCache.Value = -1
	return snap
}

func testAcquireWrite(writeBufferSize, from, to int64, cq, snap *commonQ) error {
	actualWriteSize := to - from
	if actualWriteSize == 0 && cq.failedWrites.Value != snap.failedWrites.Value+1 {
		return errors.New(fmt.Sprintf("failedWrites not incremented. Expected %d, found %d", snap.failedWrites.Value+1, cq.failedWrites.Value))
	}
	if actualWriteSize > writeBufferSize {
		return errors.New(fmt.Sprintf("Actual write size (%d) larger than requested buffer size (%d)", actualWriteSize, writeBufferSize))
	}
	if (actualWriteSize < writeBufferSize) && (cq.write.Value+actualWriteSize) != (cq.readCache.Value+cq.size) {
		if (cq.write.Value+cq.writeSize.Value)%cq.size != 0 {
			return errors.New(fmt.Sprintf("Actual write size (%d) could have been bigger.\nsnap %s\ncq  %s", actualWriteSize, snap.String(), cq.String()))
		}
	}
	if (cq.write.Value + actualWriteSize) > (cq.readCache.Value + cq.size) {
		return errors.New(fmt.Sprintf("Actual write size (%d) overwrites unread data.\ncq %s", actualWriteSize, cq.String()))
	}
	if cq.writeSize.Value != actualWriteSize {
		return errors.New(fmt.Sprintf("cq.writeSize (%d) does not equal actual write size (%d)", cq.writeSize.Value, actualWriteSize))
	}
	if from > to {
		return errors.New(fmt.Sprintf("from (%d) is greater than to (%d)", from, to))
	}
	if from >= cq.size || from < 0 {
		return errors.New(fmt.Sprintf("from (%d) must be a valid index for an array of size %d", from, cq.size))
	}
	if to > cq.size || to < 0 {
		return errors.New(fmt.Sprintf("to (%d) must be a valid index for an array of size %d", to, cq.size))
	}
	return nil
}

func testReleaseStoredWrite(cq, snap *commonQ) error {
	if cq.writeSize.Value != 0 {
		return errors.New(fmt.Sprintf("cq.writeSize was not reset to 0, %d found instead", cq.writeSize.Value))
	}
	if cq.write.Value != snap.write.Value+snap.writeSize.Value {
		return errors.New(fmt.Sprintf("write has not been advanced by the correct amount.\nsnap %s\ncq  %s", snap.String(), cq.String()))
	}
	return nil
}

func testAcquireRead(readBufferSize, from, to int64, cq, snap *commonQ) error {
	actualReadSize := to - from
	if actualReadSize == 0 && cq.failedReads.Value != snap.failedReads.Value+1 {
		return errors.New(fmt.Sprintf("failedReads not incremented. Expected %d, found %d", snap.failedReads.Value+1, cq.failedReads.Value))
	}
	if actualReadSize > readBufferSize {
		return errors.New(fmt.Sprintf("Actual read size (%d) larger than requested buffer size (%d)", actualReadSize, readBufferSize))
	}
	if (actualReadSize < readBufferSize) && (cq.read.Value+actualReadSize) != (cq.writeCache.Value) {
		if (cq.read.Value+cq.readSize.Value)%cq.size != 0 {
			return errors.New(fmt.Sprintf("Actual read size (%d) could have been bigger.\nsnap %s\ncq  %s", actualReadSize, snap.String(), cq.String()))
		}
	}
	if (cq.read.Value + actualReadSize) > cq.writeCache.Value {
		return errors.New(fmt.Sprintf("Actual read size (%d) reads past write position (%d).\ncq %s", actualReadSize, cq.write.Value, cq.String()))
	}
	if cq.readSize.Value != actualReadSize {

		return errors.New(fmt.Sprintf("cq.readSize (%d) does not equal actual read size (%d)", cq.readSize.Value, actualReadSize))
	}
	if from > to {
		return errors.New(fmt.Sprintf("from (%d) is greater than to (%d)", from, to))
	}
	if from >= cq.size || from < 0 {
		return errors.New(fmt.Sprintf("from (%d) must be a valid index for an array of size %d", from, cq.size))
	}
	if to > cq.size || to < 0 {
		return errors.New(fmt.Sprintf("to (%d) must be a valid index for an array of size %d", to, cq.size))
	}
	return nil
}

func testReleaseStoredRead(cq, snap *commonQ) error {
	if cq.readSize.Value != 0 {
		return errors.New(fmt.Sprintf("cq.readSize was not reset to 0, %d found instead", cq.readSize.Value))
	}
	if cq.read.Value != snap.read.Value+snap.readSize.Value {
		return errors.New(fmt.Sprintf("read has not been advanced by the correct amount.\nsnap %s\ncq   %s", snap.String(), cq.String()))
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
		snap := copyForWrite(cq)
		wfrom, wto := cq.acquireWrite(writeSize)
		if err := testAcquireWrite(writeSize, wfrom, wto, cq, snap); err != nil {
			t.Error(err.Error())
			return
		}
		snap = copyForWrite(cq)
		cq.releaseStoredWrite()
		if err := testReleaseStoredWrite(cq, snap); err != nil {
			t.Error(err.Error())
			return
		}
		// read
		snap = copyForRead(cq)
		rfrom, rto := cq.acquireRead(readSize)
		if err := testAcquireRead(readSize, rfrom, rto, cq, snap); err != nil {
			t.Error(err.Error())
			return
		}
		snap = copyForRead(cq)
		cq.releaseStoredRead()
		if err := testReleaseStoredRead(cq, snap); err != nil {
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
			snap := copyForWrite(cq)
			wfrom, wto := cq.acquireWrite(writeSize)
			if err := testAcquireWrite(writeSize, wfrom, wto, cq, snap); err != nil {
				t.Error(err.Error())
				return
			}
			snap = copyForWrite(cq)
			cq.releaseStoredWrite()
			if err := testReleaseStoredWrite(cq, snap); err != nil {
				t.Error(err.Error())
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
			snap := copyForRead(cq)
			rfrom, rto := cq.acquireRead(readSize)
			if err := testAcquireRead(readSize, rfrom, rto, cq, snap); err != nil {
				t.Error(err.Error())
				return
			}
			snap = copyForRead(cq)
			cq.releaseStoredRead()
			if err := testReleaseStoredRead(cq, snap); err != nil {
				t.Error(err.Error())
				return
			}
			i += (rto - rfrom)
		}
	}(&cqs)
	<-end
	<-end
}
