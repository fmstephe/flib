// +build go1.4 go1.5

package fatomic

//go:nosplit
func LazyStore(addr *int64, val int64) {
	nop()
	*addr = val
}

// Sacrificial inline method
// prevents inlining of methods will call it
// This, in turn, prevents reordering
// See runtime/atomic_amd64x.go
func nop() {}
