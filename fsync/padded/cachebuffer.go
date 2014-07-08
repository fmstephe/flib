package padded

type CacheBuffer struct {
	Bytes [cacheLineBytes * 2]byte
}
