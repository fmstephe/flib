package layout

import (
	"unsafe"
)

func CacheProtectedBytes(size int) []byte {
	b := make([]byte, cacheBytes+size+cacheBytes)
	return b[cacheBytes : size+cacheBytes]
}

func CacheProtectedPointers(size int) []unsafe.Pointer {
	b := make([]unsafe.Pointer, cacheBytes+size+cacheBytes)
	return b[cacheBytes : size+cacheBytes]
}
