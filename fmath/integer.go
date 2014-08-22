package fmath

func PowerOfTwo(val int64) bool {
	return val > 0 && val&(val-1) == 0
}

// NB: Only valid if math.MinInt64 <= x-y <= math.MaxInt64
// This is valid for these queues because x and y will always be positive
func Min(x, y int64) int64 {
	return y + ((x - y) & ((x - y) >> 63))
}

//TODO complete me
func Max(x, y int64) int64 {
	panic("Not yet implemented")
}
