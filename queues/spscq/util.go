package spscq

func powerOfTwo(val int64) bool {
	return val >= 0 && val&(val-1) == 0
}

// NB: Only valid if math.MinInt64 <= x-y <= math.MaxInt64
// This is valid for these queues because x and y will always be positive
func min(x, y int64) int64 {
	return y + ((x - y) & ((x - y) >> 63))
}
