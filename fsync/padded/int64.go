package padded

type Int64 struct {
	before [CacheLineBytes - 8]byte
	Value  int64
	after  [CacheLineBytes]byte
}
