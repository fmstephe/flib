// +build go1.6
// +build amd64

package fatomic

//go:nosplit
//go:noinline
func LazyStore(addr *int64, val int64) {
	*addr = val
}
