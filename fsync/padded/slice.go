// Copyright 2016 Francis Stephens. All rights reserved.
// Use of this source code is governed by a BSD
// license which can be found in LICENSE.txt

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
