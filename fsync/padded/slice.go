package padded

import (
	"unsafe"
)

func ByteSlice(size int) []byte {
	b := make([]byte, CacheLineBytes+size+CacheLineBytes)
	return b[CacheLineBytes : size+CacheLineBytes]
}

func PointerSlice(size int) []unsafe.Pointer {
	b := make([]unsafe.Pointer, CacheLineBytes+size+CacheLineBytes)
	return b[CacheLineBytes : size+CacheLineBytes]
}
