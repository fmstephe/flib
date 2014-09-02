package spscq

import (
	"testing"
	"github.com/fmstephe/flib/fmath"
)

func TestNewCommonQPowerOf2(t *testing.T) {
	for size := int64(1); size > 0; size *= 2 {
		newCommonQ(size)
	}
}

func TestNewCommonQNotPowerOf2(t *testing.T) {
	for size := int64(1); size < 10 * 1000; size++ {
		if !fmath.PowerOfTwo(size) {
			makeBadQ(size, t)
		}
	}
}

func makeBadQ(size int64, t *testing.T) {
	defer func(s int64) {
		if err := recover(); err == nil {
			t.Errorf("No error detected for size %d", s)
		}
	}(size)
	newCommonQ(size)
}
