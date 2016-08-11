// Copyright 2016 Francis Stephens. All rights reserved.
// Use of this source code is governed by a BSD
// license which can be found in LICENSE.txt

package padded

type CacheBuffer struct {
	Bytes [CacheLineBytes * 2]byte
}
