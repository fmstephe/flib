package fmath

import (
	"math/rand"
	"testing"
)

func TestUIGTLow(t *testing.T) {
	for x := uint64(0); x < 1000*1000; x += 1003 {
		for y := uint64(0); y <= x; y += 503 {
			testComparisons(x, y, t)
			testComparisons(y, x, t)
		}
	}
}

func TestUIGTHigh(t *testing.T) {
	for i := uint64(0); i < 1000*1000; i += 1003 {
		for j := uint64(0); j <= i; j += 503 {
			x := (0 - j) >> 1
			y := (0 - i) >> 1
			testComparisons(x, y, t)
			testComparisons(y, x, t)
		}
	}
}

func TestUIGT(t *testing.T) {
	r := rand.New(rand.NewSource(1))
	for i := 0; i < 10*1000*1000; i++ {
		x := uint64(r.Int63())
		y := uint64(r.Int63())
		testComparisons(x, y, t)
		testComparisons(y, x, t)
	}
}

func TestUINot(t *testing.T) {
	if UINot(0) != 1 {
		t.Errorf("UINot(0) returns %d", UINot(0))
	}
	if UINot(1) != 0 {
		t.Errorf("UINot(1) returns %d", UINot(1))
	}
}

func testComparisons(x, y uint64, t *testing.T) {
	if int64(x) < 0 || int64(y) < 0 {
		t.Fatalf("Cannot test numbers with highest order bit set %X, %X", x, y)
	}
	testGreater(x, y, t)
	testLess(x, y, t)
	testEqual(x, y, t)
}

// TODO is there a way to auto-generate these verbose tests?
func testGreater(x, y uint64, t *testing.T) {
	if x > y {
		resultGT := UIGT(x, y)
		if resultGT != 1 {
			t.Errorf("uint64 %d > %d but resultGT %d", x, y, resultGT)
		}
		resultNGT := UIGT(y, x)
		if resultNGT != 0 {
			t.Errorf("uint64 %x > %x but resultNGT %x", x, y, resultNGT)
		}
		resultGTE := UIGTE(x, y)
		if resultGTE != 1 {
			t.Errorf("uint64 %d > %d but resultGTE %d", x, y, resultGTE)
		}
		resultNGTE := UIGTE(y, x)
		if resultNGTE != 0 {
			t.Errorf("uint64 %d > %d but resultNGTE %d", x, y, resultNGTE)
		}
		resultLT := UILT(y, x)
		if resultLT != 1 {
			t.Errorf("uint64 %d < %d but resultLT %d", y, x, resultLT)
		}
		resultNLT := UILT(x, y)
		if resultNLT != 0 {
			t.Errorf("uint64 %d < %d but resultNLT %d", y, x, resultNLT)
		}
		resultLTE := UILTE(y, x)
		if resultLTE != 1 {
			t.Errorf("uint64 %d < %d but resultLTE %d", y, x, resultLTE)
		}
		resultNLTE := UILTE(x, y)
		if resultNLTE != 0 {
			t.Errorf("uint64 %d < %d but resultNLTE %d", y, x, resultNLTE)
		}
	}
}

func testLess(x, y uint64, t *testing.T) {
	if x < y {
		resultGT := UIGT(x, y)
		if resultGT != 0 {
			t.Errorf("uint64 %d < %d but resultGT %d", x, y, resultGT)
		}
		resultNGT := UIGT(y, x)
		if resultNGT != 1 {
			t.Errorf("uint64 %d < %d but resultNGT %d", x, y, resultNGT)
		}
		resultGTE := UIGTE(x, y)
		if resultGTE != 0 {
			t.Errorf("uint64 %d < %d but resultGTE %d", x, y, resultGTE)
		}
		resultNGTE := UIGTE(y, x)
		if resultNGTE != 1 {
			t.Errorf("uint64 %d < %d but resultNGTE %d", x, y, resultNGTE)
		}
		resultLT := UILT(y, x)
		if resultLT != 0 {
			t.Errorf("uint64 %d > %d but resultLT %d", y, x, resultLT)
		}
		resultNLT := UILT(x, y)
		if resultNLT != 1 {
			t.Errorf("uint64 %d > %d but resultNLT %d", y, x, resultNLT)
		}
		resultLTE := UILTE(y, x)
		if resultLTE != 0 {
			t.Errorf("uint64 %d > %d but resultLTE %d", y, x, resultLTE)
		}
		resultNLTE := UILTE(x, y)
		if resultNLTE != 1 {
			t.Errorf("uint64 %d > %d but resultNLTE %d", y, x, resultNLTE)
		}
	}
}

func testEqual(x, y uint64, t *testing.T) {
	if x == y {
		resultGT := UIGT(x, y)
		if resultGT != 0 {
			t.Errorf("uint64 %d == %d but resultGT %d", x, y, resultGT)
		}
		resultNGT := UIGT(y, x)
		if resultNGT != 0 {
			t.Errorf("uint64 %d == %d but resultNGT %d", x, y, resultNGT)
		}
		resultGTE := UIGTE(x, y)
		if resultGTE != 1 {
			t.Errorf("uint64 %d == %d but resultGTE %d", x, y, resultGTE)
		}
		resultNGTE := UIGTE(y, x)
		if resultNGTE != 1 {
			t.Errorf("uint64 %d == %d but resultNGTE %d", x, y, resultNGTE)
		}
		resultLT := UILT(y, x)
		if resultLT != 0 {
			t.Errorf("uint64 %d == %d but resultLT %d", y, x, resultLT)
		}
		resultNLT := UILT(x, y)
		if resultNLT != 0 {
			t.Errorf("uint64 %d == %d but resultNLT %d", y, x, resultNLT)
		}
		resultLTE := UILTE(y, x)
		if resultLTE != 1 {
			t.Errorf("uint64 %d == %d but resultLTE %d", y, x, resultLTE)
		}
		resultNLTE := UILTE(x, y)
		if resultNLTE != 1 {
			t.Errorf("uint64 %d == %d but resultNLTE %d", y, x, resultNLTE)
		}
	}
}
