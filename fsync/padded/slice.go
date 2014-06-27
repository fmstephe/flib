package padded

import (
	"unsafe"
)

func ByteSlice(size int) []byte {
	b := make([]byte, cacheLineBytes+size+cacheLineBytes)
	return b[cacheLineBytes : size+cacheLineBytes]
}

func PointerSlice(size int) []unsafe.Pointer {
	b := make([]unsafe.Pointer, cacheLineBytes+size+cacheLineBytes)
	return b[cacheLineBytes : size+cacheLineBytes]
}
