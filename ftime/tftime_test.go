package ftime

import (
	"testing"
)

func TestTest(t *testing.T) {
	println(Counter())
	println(IsCounterSteady())
	println(IsCounterSMPMonotonic())
}
