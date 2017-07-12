package spscq

import (
	"fmt"
	"reflect"
	"testing"
	"unsafe"
)

func recoverIndices(subSlice, superSlice []unsafe.Pointer) (from, to int64) {
	// NB: These two variable initialisations must be on the same line
	// otherwise the garbage collector may move the slices between
	// the two initialisations, producing nonsense results
	subHeader, superHeader := (*reflect.SliceHeader)(unsafe.Pointer(&subSlice)), (*reflect.SliceHeader)(unsafe.Pointer(&superSlice))
	ptrSize := unsafe.Sizeof(uintptr(0))
	subStart := subHeader.Data
	subEnd := subStart + (uintptr(subHeader.Len) * ptrSize)
	superStart := superHeader.Data
	superEnd := superStart + (uintptr(superHeader.Len) * ptrSize)
	if subStart < superStart {
		panic(fmt.Sprintf("subSlice (%d) starts at a lower memory address than superSlice (%d)", subStart, superStart))
	}
	if subStart > superEnd {
		panic(fmt.Sprintf("subSlice (%d) ends at a larger memory address than superSlice (%d)", subEnd, superEnd))
	}
	from = int64((subStart - superStart) / ptrSize)
	to = int64((subEnd - superStart) / ptrSize)
	return from, to
}

func TestRecoverIndices(t *testing.T) {
	for i := 0; i < 1000; i++ {
		testRecoverIndices(t, make([]unsafe.Pointer, i))
	}
}

func testRecoverIndices(t *testing.T, superSlice []unsafe.Pointer) {
	for from := range superSlice {
		for to := from; to < len(superSlice); to++ {
			subSlice := superSlice[from:to]
			rFrom, rTo := recoverIndices(subSlice, superSlice)
			if rFrom != int64(from) || rTo != int64(to) {
				t.Errorf("Bad indices returned, expected from:%d to:%d Got from:%d to:%d", from, to, rFrom, rTo)
				return
			}
		}
	}
}
