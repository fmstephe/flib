package funsafe

import (
	"reflect"
	"testing"
	"unsafe"
)

func TestStringToBytes(t *testing.T) {
	s := "string"
	b := StringToBytes(s)
	// Should have the same length
	if len(s) != len(b) {
		t.Errorf("Converted bytes have different length (%d) than the string (%d)", len(b), len(s))
	}
	if len(s) != cap(b) {
		t.Errorf("Converted bytes have capacity (%d) beyond the length of string (%d)", cap(b), len(s))
	}
	// Should have same content
	if s != string(b) {
		t.Errorf("Converted bytes has different value %q than the string %q", string(b), s)
	}
	// Should point to the same data in memory
	sData := (*(*reflect.StringHeader)(unsafe.Pointer(&s))).Data
	bData := (*(*reflect.SliceHeader)(unsafe.Pointer(&b))).Data
	if sData != bData {
		t.Errorf("Converted bytes points to different data %d than the string %d", sData, bData)
	}
}

func TestBytesToString(t *testing.T) {
	b := []byte("bytes!")
	s := BytesToString(b)
	// Should have the same length
	if len(s) != len(b) {
		t.Errorf("Converted string has a different length (%d) than the bytes (%d)", len(s), len(b))
	}
	// Should have same content
	if s != string(b) {
		t.Errorf("Converted string has a different value %q than the bytes %q", s, string(b))
	}
	// Should point to the same data in memory
	sData := (*(*reflect.StringHeader)(unsafe.Pointer(&s))).Data
	bData := (*(*reflect.SliceHeader)(unsafe.Pointer(&b))).Data
	if sData != bData {
		t.Errorf("Converted string points to different data %d than the bytes %d", sData, bData)
	}
}

// Check we don't access the entire byte slice's capacity
func TestBytesToString_WithUnusedBytes(t *testing.T) {
	// make a long slice of bytes
	bLongDontUse := []byte("bytes! and all these other bytes")
	// just take the first 6 characters
	b := bLongDontUse[:6]
	s := BytesToString(b)
	// Should have the same length
	if len(s) != len(b) {
		t.Errorf("Converted string has a different length (%d) than the bytes (%d)", len(s), len(b))
	}
	// Should have same content
	if s != string(b) {
		t.Errorf("Converted string has a different value %q than the bytes %q", s, string(b))
	}
	// Should point to the same data in memory
	sData := (*(*reflect.StringHeader)(unsafe.Pointer(&s))).Data
	bData := (*(*reflect.SliceHeader)(unsafe.Pointer(&b))).Data
	if sData != bData {
		t.Errorf("Converted string points to different data %d than the bytes %d", sData, bData)
	}
}

func TestStringHeadersCompatible(t *testing.T) {
	// Check to make sure string header is what reflect thinks it is.
	// They should be the same except for the type of the data field.
	if unsafe.Sizeof(stringHeader{}) != unsafe.Sizeof(reflect.StringHeader{}) {
		t.Errorf("stringHeader layout has changed ours %#v theirs %#v", stringHeader{}, reflect.StringHeader{})
	}
	x := stringHeader{}
	y := reflect.StringHeader{}
	x.data = unsafe.Pointer(y.Data)
	y.Data = uintptr(x.data)
	x.stringLen = y.Len
	y.Len = x.stringLen
	// If we can do all of that then the two structs are compatible
}

func TestSliceHeadersCompatible(t *testing.T) {
	// Check to make sure string header is what reflect thinks it is.
	// They should be the same except for the type of the data field.
	if unsafe.Sizeof(sliceHeader{}) != unsafe.Sizeof(reflect.SliceHeader{}) {
		t.Errorf("sliceHeader layout has changed ours %#v theirs %#v", sliceHeader{}, reflect.SliceHeader{})
	}
	x := sliceHeader{}
	y := reflect.SliceHeader{}
	x.data = unsafe.Pointer(y.Data)
	y.Data = uintptr(x.data)
	x.sliceLen = y.Len
	y.Len = x.sliceLen
	x.sliceCap = y.Cap
	y.Cap = x.sliceCap
	// If we can do all of that then the two structs are compatible
}
