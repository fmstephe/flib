package spscq

import (
	"errors"
	"fmt"
	"github.com/fmstephe/flib/fmath"
	"math/rand"
	"testing"
)

// Test that we can call newCommonQ(...) for every power of 2 in an int64
func TestNewCommonQPowerOf2(t *testing.T) {
	for size := int64(1); size <= maxSize; size *= 2 {
		_, err := newCommonQ(size)
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
	_, err := newCommonQ(size)
	if err == nil {
		t.Errorf("No error detected for size %d", size)
	}
}

func testAcquireWrite(writeBufferSize, from, to int64, cq, snap commonQ) error {
	actualWriteSize := to - from
	if actualWriteSize == 0 && cq.failedWrites.Value != snap.failedWrites.Value+1 {
		msg := "failedWrites not incremented. Expected %d, found %d"
		return errors.New(fmt.Sprintf(msg, snap.failedWrites.Value+1, cq.failedWrites.Value))
	}
	if actualWriteSize > writeBufferSize {
		msg := "Actual write size (%d) larger than requested buffer size (%d)"
		return errors.New(fmt.Sprintf(msg, actualWriteSize, writeBufferSize))
	}
	if (actualWriteSize < writeBufferSize) && (cq.write.Value+actualWriteSize) != (cq.read.Value+cq.size) {
		msg := "Actual write size (%d) could have been bigger.\nsnap %s\ncq  %s"
		return errors.New(fmt.Sprintf(msg, actualWriteSize, snap.String(), cq.String()))
	}
	if (cq.write.Value + actualWriteSize) > (cq.read.Value + cq.size) {
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

func testReleaseWrite(cq, snap commonQ) error {
	if cq.writeSize.Value != 0 {
		return errors.New(fmt.Sprintf("cq.writeSize was not reset to 0, %d found instead", cq.writeSize))
	}
	if cq.write.Value != snap.write.Value+snap.writeSize.Value {
		return errors.New(fmt.Sprintf("write has not been advanced by the correct amount.\nsnap %s\ncq  %s", snap.String(), cq.String()))
	}
	return nil
}

func testAcquireRead(readBufferSize, from, to int64, cq, snap commonQ) error {
	actualReadSize := to - from
	if actualReadSize == 0 && cq.failedReads.Value != snap.failedReads.Value+1 {
		msg := "failedReads not incremented. Expected %d, found %d"
		return errors.New(fmt.Sprintf(msg, snap.failedReads.Value+1, cq.failedReads.Value))
	}
	if actualReadSize > readBufferSize {
		msg := "Actual read size (%d) larger than requested buffer size (%d)"
		return errors.New(fmt.Sprintf(msg, actualReadSize, readBufferSize))
	}
	if (actualReadSize < readBufferSize) && (cq.read.Value+actualReadSize) != (cq.write.Value) {
		msg := "Actual read size (%d) could have been bigger.\nsnap %s\ncq  %s"
		return errors.New(fmt.Sprintf(msg, actualReadSize, snap.String(), cq.String()))
	}
	if (cq.read.Value + actualReadSize) > cq.write.Value {
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

func testReleaseRead(cq, snap commonQ) error {
	if cq.readSize.Value != 0 {
		return errors.New(fmt.Sprintf("cq.readSize was not reset to 0, %d found instead", cq.readSize))
	}
	if cq.read.Value != snap.read.Value+snap.readSize.Value {
		return errors.New(fmt.Sprintf("read has not been advanced by the correct amount.\nsnap %s\ncq   %s", snap.String(), cq.String()))
	}
	return nil
}

func TestEvenReadWrites(t *testing.T) {
	rand.Seed(1)
	for i := uint64(0); i <= 41; i++ {
		size := int64(1 << i)
		cq, err := newCommonQ(size)
		if err != nil {
			t.Error(err.Error())
			continue
		}
		bufferSize := fmath.Max(size/128, 1)
		for j := int64(0); j < 1024; j++ {
			// write
			snap := cq
			wfrom, wto := cq.acquireWrite(bufferSize)
			if err := testAcquireWrite(bufferSize, wfrom, wto, cq, snap); err != nil {
				t.Error(err.Error())
				return
			}
			snap = cq
			cq.ReleaseWrite()
			if err := testReleaseWrite(cq, snap); err != nil {
				t.Error(err.Error())
				return
			}
			// read
			snap = cq
			rfrom, rto := cq.acquireRead(bufferSize)
			if err := testAcquireRead(bufferSize, rfrom, rto, cq, snap); err != nil {
				t.Error(err.Error())
				return
			}
			snap = cq
			cq.ReleaseRead()
			if err := testReleaseRead(cq, snap); err != nil {
				t.Error(err.Error())
				return
			}
		}
	}
}
