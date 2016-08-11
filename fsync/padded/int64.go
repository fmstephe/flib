// Copyright 2016 Francis Stephens. All rights reserved.
// Use of this source code is governed by a BSD
// license which can be found in LICENSE.txt

package padded

type Int64 struct {
	before [CacheLineBytes - 8]byte
	Value  int64
	after  [CacheLineBytes]byte
}
