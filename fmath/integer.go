package fmath

// Returns true if val is a power of two, otherwise returns false
func PowerOfTwo(val int64) bool {
	return val > 0 && val&(val-1) == 0
}

// Returns x if x < y, otherwise returns y
//
// NB: Only valid if math.MinInt64 <= x-y <= math.MaxInt64
// In particular, always valid if both arguments are positive
func Min(x, y int64) int64 {
	return y + ((x - y) & ((x - y) >> 63))
}

// Returns x if x > y, otherwise returns y
//
// NB: Only valid if math.MinInt64 <= x-y <= math.MaxInt64
// In particular, always valid if both arguments are positive
func Max(x, y int64) int64 {
	return x ^ ((x ^ y) & ((x - y) >> 63))
}

// Combines two int32 values into a single int64
// high occupies bits 32-63
// low occupies bits 0-31
func CombineInt32(high, low int32) int64 {
	high64 := int64(uint32(high)) << 32
	low64 := int64(uint32(low))
	return high64 | low64
}

// Returns the highest 32 bits of an int64
func HighInt32(whole int64) int32 {
	return int32(whole >> 32)
}

// Returns the lowest 32 bits of an int64
func LowInt32(whole int64) int32 {
	return int32(whole)
}
