// Copyright 2016 Francis Stephens. All rights reserved.
// Use of this source code is governed by a BSD
// license which can be found in LICENSE.txt

package fmath

import (
	"testing"
)

func TestUI8All(t *testing.T) {
	for x := uint8(0); x <= 127 && x >= 0; x++ {
		for y := uint8(0); y <= 127; y++ {
			testComparisonsUI8(x, y, t)
			testComparisonsUI8(y, x, t)
		}
	}
}

func TestUI8Not(t *testing.T) {
	if UI8Not(0) != 1 {
		t.Errorf("UI8Not(0) returns %d", UI8Not(0))
	}
	if UI8Not(1) != 0 {
		t.Errorf("UI8Not(1) returns %d", UI8Not(1))
	}
}

func testComparisonsUI8(x, y uint8, t *testing.T) {
	if int64(x) < 0 || int64(y) < 0 {
		t.Fatalf("Cannot test numbers with highest order bit set %X, %X", x, y)
	}
	testGreaterUI8(x, y, t)
	testLessUI8(x, y, t)
	testEqualUI8(x, y, t)
}

// TODO is there a way to auto-generate these verbose tests?
func testGreaterUI8(x, y uint8, t *testing.T) {
	if x > y {
		resultGT := UI8GT(x, y)
		if resultGT != 1 {
			println(x, y, y-x)
			t.Errorf("uint8 %d > %d but resultGT %d", x, y, resultGT)
		}
		resultNGT := UI8GT(y, x)
		if resultNGT != 0 {
			t.Errorf("uint8 %x > %x but resultNGT %x", x, y, resultNGT)
		}
		resultGTE := UI8GTE(x, y)
		if resultGTE != 1 {
			t.Errorf("uint8 %d > %d but resultGTE %d", x, y, resultGTE)
		}
		resultNGTE := UI8GTE(y, x)
		if resultNGTE != 0 {
			t.Errorf("uint8 %d > %d but resultNGTE %d", x, y, resultNGTE)
		}
		resultLT := UI8LT(y, x)
		if resultLT != 1 {
			t.Errorf("uint8 %d < %d but resultLT %d", y, x, resultLT)
		}
		resultNLT := UI8LT(x, y)
		if resultNLT != 0 {
			t.Errorf("uint8 %d < %d but resultNLT %d", y, x, resultNLT)
		}
		resultLTE := UI8LTE(y, x)
		if resultLTE != 1 {
			t.Errorf("uint8 %d < %d but resultLTE %d", y, x, resultLTE)
		}
		resultNLTE := UI8LTE(x, y)
		if resultNLTE != 0 {
			t.Errorf("uint8 %d < %d but resultNLTE %d", y, x, resultNLTE)
		}
	}
}

func testLessUI8(x, y uint8, t *testing.T) {
	if x < y {
		resultGT := UI8GT(x, y)
		if resultGT != 0 {
			t.Errorf("uint8 %d < %d but resultGT %d", x, y, resultGT)
		}
		resultNGT := UI8GT(y, x)
		if resultNGT != 1 {
			t.Errorf("uint8 %d < %d but resultNGT %d", x, y, resultNGT)
		}
		resultGTE := UI8GTE(x, y)
		if resultGTE != 0 {
			t.Errorf("uint8 %d < %d but resultGTE %d", x, y, resultGTE)
		}
		resultNGTE := UI8GTE(y, x)
		if resultNGTE != 1 {
			t.Errorf("uint8 %d < %d but resultNGTE %d", x, y, resultNGTE)
		}
		resultLT := UI8LT(y, x)
		if resultLT != 0 {
			t.Errorf("uint8 %d > %d but resultLT %d", y, x, resultLT)
		}
		resultNLT := UI8LT(x, y)
		if resultNLT != 1 {
			t.Errorf("uint8 %d > %d but resultNLT %d", y, x, resultNLT)
		}
		resultLTE := UI8LTE(y, x)
		if resultLTE != 0 {
			t.Errorf("uint8 %d > %d but resultLTE %d", y, x, resultLTE)
		}
		resultNLTE := UI8LTE(x, y)
		if resultNLTE != 1 {
			t.Errorf("uint8 %d > %d but resultNLTE %d", y, x, resultNLTE)
		}
	}
}

func testEqualUI8(x, y uint8, t *testing.T) {
	if x == y {
		resultGT := UI8GT(x, y)
		if resultGT != 0 {
			t.Errorf("uint8 %d == %d but resultGT %d", x, y, resultGT)
		}
		resultNGT := UI8GT(y, x)
		if resultNGT != 0 {
			t.Errorf("uint8 %d == %d but resultNGT %d", x, y, resultNGT)
		}
		resultGTE := UI8GTE(x, y)
		if resultGTE != 1 {
			t.Errorf("uint8 %d == %d but resultGTE %d", x, y, resultGTE)
		}
		resultNGTE := UI8GTE(y, x)
		if resultNGTE != 1 {
			t.Errorf("uint8 %d == %d but resultNGTE %d", x, y, resultNGTE)
		}
		resultLT := UI8LT(y, x)
		if resultLT != 0 {
			t.Errorf("uint8 %d == %d but resultLT %d", y, x, resultLT)
		}
		resultNLT := UI8LT(x, y)
		if resultNLT != 0 {
			t.Errorf("uint8 %d == %d but resultNLT %d", y, x, resultNLT)
		}
		resultLTE := UI8LTE(y, x)
		if resultLTE != 1 {
			t.Errorf("uint8 %d == %d but resultLTE %d", y, x, resultLTE)
		}
		resultNLTE := UI8LTE(x, y)
		if resultNLTE != 1 {
			t.Errorf("uint8 %d == %d but resultNLTE %d", y, x, resultNLTE)
		}
	}
}
