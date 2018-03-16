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

func testReadOnlyAcquireReleaser(from, to int64, before, after acquireReleaser) error {
	if before.name != after.name {
		return fmt.Errorf("before.name (%s) does not equal after.name (%s)", before.name, after.name)
	}
	if before.size != after.size {
		return fmt.Errorf("before.size (%d) does not equal after.size (%d)", before.size, after.size)
	}
	if before.mask != after.mask {
		return fmt.Errorf("before.mask (%d) does not equal after.mask (%d)", before.mask, after.mask)
	}
	if before.pause != after.pause {
		return fmt.Errorf("before.pause (%d) does not equal after.pause (%d)", before.pause, after.pause)
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
	return nil
}

func testRelease(before, after acquireReleaser) error {
	if after.released != before.released+before.unreleased {
		return errors.New(fmt.Sprintf("released has not been advanced by the correct amount.\nbefore %s\nafter  %s", &before, &after))
	}
	return nil
}

// Test that we can call newCommonQ(...) for every power of 2 in an int64
func TestNewCommonQPowerOf2(t *testing.T) {
	for size := int64(1); size <= maxSize; size *= 2 {
		_, err := newPointerCommonQ(size, 0)
		if err != nil {
			t.Errorf("Error found for size %d", size)
		}
	}
}

// Test that we can't call newCommonQ(...) with a non-power of 2 size
func TestNewCommonQNotPowerOf2(t *testing.T) {
	for size := int64(1); size < 10*1000; size++ {
		if !fmath.PowerOfTwo(size) {
			_, err := newPointerCommonQ(size, 0)
			if err == nil {
				t.Errorf("No error detected for size %d", size)
			}
		}
	}
}
