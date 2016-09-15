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
			makeBadQ(size, t)
		}
	}
}

func makeBadQ(size int64, t *testing.T) {
	_, err := newCommonQ(size, 0)
	if err == nil {
		t.Errorf("No error detected for size %d", size)
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
		msg := "failedWrites not incremented. Expected %d, found %d"
		return errors.New(fmt.Sprintf(msg, snap.failedWrites.Value+1, cq.failedWrites.Value))
	}
	if actualWriteSize > writeBufferSize {
		msg := "Actual write size (%d) larger than requested buffer size (%d)"
		return errors.New(fmt.Sprintf(msg, actualWriteSize, writeBufferSize))
	}
	if (actualWriteSize < writeBufferSize) && (cq.write.Value+actualWriteSize) != (cq.readCache.Value+cq.size) {
		if (cq.write.Value+cq.writeSize.Value)%cq.size != 0 {
			msg := "Actual write size (%d) could have been bigger.\nsnap %s\ncq  %s"
			return errors.New(fmt.Sprintf(msg, actualWriteSize, snap.String(), cq.String()))
		}
	}
	if (cq.write.Value + actualWriteSize) > (cq.readCache.Value + cq.size) {
		msg := "Actual write size (%d) overwrites unread data.\ncq %s"
		return errors.New(fmt.Sprintf(msg, actualWriteSize, cq.String()))
	}
	if cq.writeSize.Value != actualWriteSize {
		msg := "cq.writeSize does not equal actual write size"
		return errors.New(fmt.Sprintf(msg, cq.writeSize, actualWriteSize))
	}
	if from > to {
		msg := "from (%d) is greater than to (%d)"
		return errors.New(fmt.Sprintf(msg, from, to))
	}
	if from >= cq.size || from < 0 {
		msg := "from (%d) must be a valid index for an array of size %d"
		return errors.New(fmt.Sprintf(msg, from, cq.size))
	}
	if to > cq.size || to < 0 {
		msg := "to (%d) must be a valid index for an array of size %d"
		return errors.New(fmt.Sprintf(msg, to, cq.size))
	}
	return nil
}

func testReleaseWrite(cq, snap *commonQ) error {
	if cq.writeSize.Value != 0 {
		return errors.New(fmt.Sprintf("cq.writeSize was not reset to 0, %d found instead", cq.writeSize))
	}
	if cq.write.Value != snap.write.Value+snap.writeSize.Value {
		return errors.New(fmt.Sprintf("write has not been advanced by the correct amount.\nsnap %s\ncq  %s", snap.String(), cq.String()))
	}
	return nil
}

func testAcquireRead(readBufferSize, from, to int64, cq, snap *commonQ) error {
	actualReadSize := to - from
	if actualReadSize == 0 && cq.failedReads.Value != snap.failedReads.Value+1 {
		msg := "failedReads not incremented. Expected %d, found %d"
		return errors.New(fmt.Sprintf(msg, snap.failedReads.Value+1, cq.failedReads.Value))
	}
	if actualReadSize > readBufferSize {
		msg := "Actual read size (%d) larger than requested buffer size (%d)"
		return errors.New(fmt.Sprintf(msg, actualReadSize, readBufferSize))
	}
	if (actualReadSize < readBufferSize) && (cq.read.Value+actualReadSize) != (cq.writeCache.Value) {
		if (cq.read.Value+cq.readSize.Value)%cq.size != 0 {
			msg := "Actual read size (%d) could have been bigger.\nsnap %s\ncq  %s"
			return errors.New(fmt.Sprintf(msg, actualReadSize, snap.String(), cq.String()))
		}
	}
	if (cq.read.Value + actualReadSize) > cq.writeCache.Value {
		msg := "Actual read size (%d) reads past write position (%d).\ncq %s"
		return errors.New(fmt.Sprintf(msg, actualReadSize, cq.write.Value, cq.String()))
	}
	if cq.readSize.Value != actualReadSize {
		msg := "cq.readSize does not equal actual read size"
		return errors.New(fmt.Sprintf(msg, cq.readSize, actualReadSize))
	}
	if from > to {
		msg := "from (%d) is greater than to (%d)"
		return errors.New(fmt.Sprintf(msg, from, to))
	}
	if from >= cq.size || from < 0 {
		msg := "from (%d) must be a valid index for an array of size %d"
		return errors.New(fmt.Sprintf(msg, from, cq.size))
	}
	if to > cq.size || to < 0 {
		msg := "to (%d) must be a valid index for an array of size %d"
		return errors.New(fmt.Sprintf(msg, to, cq.size))
	}
	return nil
}

func testReleaseRead(cq, snap *commonQ) error {
	if cq.readSize.Value != 0 {
		return errors.New(fmt.Sprintf("cq.readSize was not reset to 0, %d found instead", cq.readSize))
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
		cq.ReleaseWrite()
		if err := testReleaseWrite(cq, snap); err != nil {
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
		cq.ReleaseRead()
		if err := testReleaseRead(cq, snap); err != nil {
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
			cq.ReleaseWrite()
			if err := testReleaseWrite(cq, snap); err != nil {
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
			cq.ReleaseRead()
			if err := testReleaseRead(cq, snap); err != nil {
				t.Error(err.Error())
				return
			}
			i += (rto - rfrom)
		}
	}(&cqs)
	<-end
	<-end
}
