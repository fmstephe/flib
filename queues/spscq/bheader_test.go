package spscq

import (
	"encoding/binary"
	"testing"
)

// These tests are intended to provide a rough guide
// to the performance cost of writing an int64 to a
// byte slice using unsafe. These benchmarks are very
// micro and therefore do not include cache misses or
// other confounding factors. But they are definitely
// interesting.

func BenchmarkComparisonWriteHeader(b *testing.B) {
	x := int64(0)
	for i := 0; i < b.N; i++ {
		x = int64(i)
	}
	_ = x
}

func BenchmarkComparisonReadHeader(b *testing.B) {
	x := int64(0)
	y := int64(0)
	for i := 0; i < b.N; i++ {
		y = x
	}
	_ = x
	_ = y
}

func BenchmarkWriteHeader(b *testing.B) {
	x := make([]byte, 8)
	for i := 0; i < b.N; i++ {
		writeHeader(x, 0, int64(i))
	}
}

func BenchmarkReadHeader(b *testing.B) {
	y := int64(0)
	x := make([]byte, 8)
	writeHeader(x, 0, 64)
	for i := 0; i < b.N; i++ {
		y = readHeader(x, 0)
	}
	_ = y
}

func BenchmarkBinaryWriteHeader(b *testing.B) {
	x := make([]byte, 8)
	for i := 0; i < b.N; i++ {
		binary.PutVarint(x, int64(i))
	}
}

func BenchmarkBinaryReadHeader(b *testing.B) {
	y := int64(0)
	x := make([]byte, 8)
	writeHeader(x, 0, 64)
	for i := 0; i < b.N; i++ {
		y, _ = binary.Varint(x)
	}
	_ = y
}
