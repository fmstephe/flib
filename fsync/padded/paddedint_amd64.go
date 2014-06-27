package padded

type Int64 struct {
	before [7]int64
	Value  int64
	after  [8]int64
}
