package fmath

import (
	"math/rand"
	"testing"

	"github.com/fmstephe/flib/fstrconv"
)

func TestMin(t *testing.T) {
	rand.Seed(1)
	for i := 0; i < 1000*1000; i++ {
		a := rand.Int63n(1000 * 1000 * 1000)
		b := rand.Int63n(1000 * 1000 * 1000)
		m := Min(a, b)
		om := simpleMin(a, b)
		if m != om {
			as := fstrconv.ItoaComma(a)
			bs := fstrconv.ItoaComma(b)
			ms := fstrconv.ItoaComma(m)
			t.Errorf("Problem with min of %s, %s - min returned %s", as, bs, ms)
		}
	}
}

func simpleMin(val1, val2 int64) int64 {
	if val1 < val2 {
		return val1
	}
	return val2
}
