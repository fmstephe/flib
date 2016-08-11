// Copyright 2016 Francis Stephens. All rights reserved.
// Use of this source code is governed by a BSD
// license which can be found in LICENSE.txt

// +build go1.6
// +build amd64

package fatomic

//go:nosplit
//go:noinline
func LazyStore(addr *int64, val int64) {
	*addr = val
}
