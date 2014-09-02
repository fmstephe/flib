package fmath

import (
	"math"
	"math/rand"
	"testing"

	"github.com/fmstephe/flib/fstrconv"
)

var allPowersOfTwo []int64

func init() {
	allPowersOfTwo = make([]int64, 1)
	allPowersOfTwo[0] = 1
	for i := int64(2); i > 0; i = i << 1 {
		allPowersOfTwo = append(allPowersOfTwo, i)
	}
}

// Test fmath.PowerOfTwo(int64) bool
func TestPowerOfTwo(t *testing.T) {
	// Test all actual powers of two
	for _, i := range allPowersOfTwo {
		checkPowerOfTwo(t, i)
	}
	// Test low numbers for power of two-ness
	for i := int64(0); i < 10*1000; i++ {
		checkPowerOfTwo(t, i)
	}
	// Test high numbers for power of two-ness
	for i := int64(math.MaxInt64); i > math.MaxInt64-(10*1000); i-- {
		checkPowerOfTwo(t, i)
	}
	// Test small negatives for power of two-ness
	for i := int64(0); i > -10*1000; i-- {
		checkPowerOfTwo(t, i)
	}
	// Test large negatives for power of two-ness
	for i := int64(math.MinInt64); i < math.MinInt64+(10*1000); i++ {
		checkPowerOfTwo(t, i)
	}
	// Test random numbers for power of two-ness
	rand.Seed(1)
	for i := 0; i < 10*1000; i++ {
		n := rand.Int63()
		checkPowerOfTwo(t, n)
	}
}

func checkPowerOfTwo(t *testing.T, i int64) {
	r := PowerOfTwo(i)
	rs := simplePowerOfTwo(i)
	if r != rs {
		t.Errorf("PowerOfTwo(%d) returns %v, while simplePowerOfTwo(%d) returns %v", i, r, i, rs)
	}
}

func simplePowerOfTwo(i int64) bool {
	for _, j := range allPowersOfTwo {
		if i == j {
			return true
		}
	}
	return false
}

// Test fmath.Min(int64,int64) int64
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

// Test with random positive int32
// CombineInt32(int32,int32) int64
// HighInt32(int64) int32
// LowInt32(int64) int32
func TestCombineInt32(t *testing.T) {
	r := rand.New(rand.NewSource(1))
	for i := 0; i < 10*1000; i++ {
		high := r.Int31()
		low := r.Int31()
		whole := CombineInt32(high, low)
		if high != HighInt32(whole) {
			t.Errorf("Expecting '%d' found '%d'", high, HighInt32(whole))
		}
		if low != LowInt32(whole) {
			t.Errorf("Expecting '%d' found '%d'", low, LowInt32(whole))
		}
	}
}

// Test with random negative int32
// CombineInt32(int32,int32) int64
// HighInt32(int64) int32
// LowInt32(int64) int32
func TestGuidFunsWithNegativeInt32(t *testing.T) {
	r := rand.New(rand.NewSource(1))
	for i := 0; i < 10*1000; i++ {
		high := -r.Int31()
		low := -r.Int31()
		whole := CombineInt32(high, low)
		if high != HighInt32(whole) {
			t.Errorf("Expecting '%d' found '%d'", high, HighInt32(whole))
		}
		if low != LowInt32(whole) {
			t.Errorf("Expecting '%d' found '%d'", low, LowInt32(whole))
		}
	}
}

// Test with random uint32, using int32 casts
// CombineInt32(int32,int32) int64
// HighInt32(int64) int32
// LowInt32(int64) int32
func TestCombineUint32(t *testing.T) {
	r := rand.New(rand.NewSource(1))
	for i := 0; i < 10*1000; i++ {
		high := r.Uint32()
		low := r.Uint32()
		whole := CombineInt32(int32(high), int32(low))
		if high != uint32(HighInt32(whole)) {
			t.Errorf("Expecting '%d' found '%d'", high, uint32(HighInt32(whole)))
		}
		if low != uint32(LowInt32(whole)) {
			t.Errorf("Expecting '%d' found '%d'", low, uint32(LowInt32(whole)))
		}
	}
}

// Test with random uint32 most significant bit set, using int32 casts
// CombineInt32(int32,int32) int64
// HighInt32(int64) int32
// LowInt32(int64) int32
func TestGuidFunsWithLargeUint32(t *testing.T) {
	r := rand.New(rand.NewSource(1))
	for i := 0; i < 10*1000; i++ {
		high := uint32(-r.Int31())
		low := uint32(-r.Int31())
		whole := CombineInt32(int32(high), int32(low))
		if high != uint32(HighInt32(whole)) {
			t.Errorf("Expecting '%d' found '%d'", high, HighInt32(whole))
		}
		if low != uint32(LowInt32(whole)) {
			t.Errorf("Expecting '%d' found '%d'", low, LowInt32(whole))
		}
	}
}
