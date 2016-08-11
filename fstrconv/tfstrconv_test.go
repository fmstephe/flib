// Copyright 2016 Francis Stephens. All rights reserved.
// Use of this source code is governed by a BSD
// license which can be found in LICENSE.txt

package fstrconv

import (
	"testing"
)

func TestTest(t *testing.T) {
	helper(0, "0", t)
	helper(1, "1", t)
	helper(2, "2", t)
	helper(3, "3", t)
	helper(7, "7", t)
	helper(8, "8", t)
	helper(10, "10", t)
	helper(33, "33", t)
	helper(99, "99", t)
	helper(100, "100", t)
	helper(123, "123", t)
	helper(999, "999", t)
	helper(1000, "1,000", t)
	helper(10000, "10,000", t)
	helper(100000, "100,000", t)
	helper(1000000, "1,000,000", t)
	helper(10000000, "10,000,000", t)
	helper(100000000, "100,000,000", t)
	helper(1000000000, "1,000,000,000", t)

	helper(0, "0", t)
	helper(0, "0", t)
	helper(1, "1", t)
	helper(2, "2", t)
	helper(3, "3", t)
	helper(7, "7", t)
	helper(8, "8", t)
	helper(10, "10", t)
	helper(33, "33", t)
	helper(99, "99", t)
	helper(100, "100", t)
	helper(123, "123", t)
	helper(999, "999", t)
	helper(1000, "1,000", t)
	helper(10000, "10,000", t)
	helper(100000, "100,000", t)
	helper(1000000, "1,000,000", t)
	helper(10000000, "10,000,000", t)
	helper(100000000, "100,000,000", t)
	helper(1000000000, "1,000,000,000", t)

	helper(-1, "-1", t)
	helper(-2, "-2", t)
	helper(-3, "-3", t)
	helper(-7, "-7", t)
	helper(-8, "-8", t)
	helper(-10, "-10", t)
	helper(-33, "-33", t)
	helper(-99, "-99", t)
	helper(-100, "-100", t)
	helper(-123, "-123", t)
	helper(-999, "-999", t)
	helper(-1000, "-1,000", t)
	helper(-10000, "-10,000", t)
	helper(-100000, "-100,000", t)
	helper(-1000000, "-1,000,000", t)
	helper(-10000000, "-10,000,000", t)
	helper(-100000000, "-100,000,000", t)
	helper(-1000000000, "-1,000,000,000", t)
}

func helper(x int64, s string, t *testing.T) {
	r := ItoaDelim(x, ',')
	if r != s {
		t.Errorf("%d not reversed properly, expecting %s got %s instead", x, s, r)
	}
}
