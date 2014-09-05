package spscq

import (
	"github.com/fmstephe/flib/fmath"
	"math/rand"
	"testing"
)

const maxSize = 1 << 42 // Represents a very big ring buffer

// Test that we can call newCommonQ(...) for every power of 2 in an int64
func TestNewCommonQPowerOf2(t *testing.T) {
	for size := int64(1); size > 0; size *= 2 {
		newCommonQ(size)
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

// commonQ.AcquireWrite(...)/commonQ.ReleaseWrite() advances write field
func TestAcquireReleaseWriteAdvancesWriteField(t *testing.T) {
	rand.Seed(1)
	for size := int64(1); size <= maxSize; size *= 2 {
		rem := size
		cq, _ := newCommonQ(size)
		for rem > 0 {
			bufferSize := rand.Int63n(rem) + 1
			cq.acquireWrite(bufferSize)
			cq.ReleaseWrite()
			rem -= bufferSize
			if cq.write.Value != size-rem {
				t.Errorf("Expecting cq.write == %d, found %d", size-rem, cq.write.Value)
			}
		}
	}
}

// commonQ.AcquireRead(...)/commonQ.ReleaseRead() advances read field
func TestAcquireReleaseReadAdvancesReadField(t *testing.T) {
	rand.Seed(1)
	for size := int64(1); size <= maxSize; size *= 2 {
		rem := size
		cq, _ := newCommonQ(size)
		cq.acquireWrite(size)
		cq.ReleaseWrite()
		for rem > 0 {
			bufferSize := rand.Int63n(rem) + 1
			cq.acquireRead(bufferSize)
			cq.ReleaseRead()
			rem -= bufferSize
			if cq.read.Value != size-rem {
				t.Errorf("Expecting cq.read == %d, found %d", size-rem, cq.read.Value)
			}
		}
	}
}

// commonQ.AcquireWrite(...)/commonQ.ReleaseWrite() does not allow write to wrap past read
func TestAcquireWriteDoesNotWrapPastRead(t *testing.T) {
	rand.Seed(1)
	for size := int64(1); size <= maxSize; size *= 2 {
		writeRand := (size / 64) + 1
		readRand := (size / 128) + 1
		cq, _ := newCommonQ(size)
		for i := 0; i < 10*1000; i++ {
			writeSize := rand.Int63n(writeRand) + 2
			readSize := rand.Int63n(readRand) + 1
			cq.acquireWrite(writeSize)
			cq.ReleaseWrite()
			cq.acquireRead(readSize)
			cq.ReleaseRead()
			if cq.write.Value-size > cq.read.Value {
				t.Errorf("Write has overlapped read. cq.write = %d, cq.read = %d, cq.size = %d", cq.write.Value, cq.read.Value, cq.size)
			}
		}
	}
}

// commonQ.AcquireWrite(...)/commonQ.ReleaseWrite() does not allow read to overtake write
func TestAcquireReadDoesNotOvertakeWrite(t *testing.T) {
	rand.Seed(1)
	for size := int64(1); size <= maxSize; size *= 2 {
		writeRand := (size / 128) + 1
		readRand := (size / 64) + 1
		cq, _ := newCommonQ(size)
		for i := 0; i < 10*1000; i++ {
			writeSize := rand.Int63n(writeRand) + 2
			readSize := rand.Int63n(readRand) + 1
			cq.acquireWrite(writeSize)
			cq.ReleaseWrite()
			cq.acquireRead(readSize)
			cq.ReleaseRead()
			if cq.read.Value > cq.write.Value {
				t.Errorf("Read has overtaken Write. cq.read = %d, cq.write = %d, cq.size = %d", cq.read.Value, cq.write.Value, cq.size)
			}
		}
	}
}
