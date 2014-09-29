package padded

type CacheBuffer struct {
	Bytes [CacheLineBytes * 2]byte
}
