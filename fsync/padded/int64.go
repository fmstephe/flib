package padded

type Int64 struct {
	before [cacheLineBytes - 8]byte
	Value  int64
	after  [cacheLineBytes]byte
}
