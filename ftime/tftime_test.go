package ftime

import (
	. "github.com/fmstephe/flib/fstrconv"
	"testing"
)

func TestCounter(t *testing.T) {
	if !IsCounterSMPMonotonic() {
		return // Very hard to guarantee anything
	}
	c := Counter()
	for i := 0; i < 1000*1000; i++ {
		newc := Counter()
		if newc < c {
			t.Errorf("Counter() values reducing. Previous %s New %s", ItoaComma(c), ItoaComma(newc))
		}
	}
}

func TestCounterSMP(t *testing.T) {
	if !IsCounterSMPMonotonic() {
		return // Very hard to guarantee anything
	}
	counterChan := make(chan int64)
	go func() {
		for i := 0; i < 1000*1000; i++ {
			counterChan <- Counter()
		}
		close(counterChan)
	}()
	go func() {
		for c := range counterChan {
			newc := Counter()
			if newc < c {
				t.Errorf("Counter() values reducing. Previous %s New %s", ItoaComma(c), ItoaComma(newc))
			}
		}
	}()
}

func TestPause(t *testing.T) {
	for i := int64(1000); i <= int64(1000*1000*1000); i *= 10 {
		c := Counter()
		Pause(i)
		c = Counter() - c
		if c < i {
			t.Errorf("Counter ticks elapsed (%s) less than asked for (%s)", ItoaComma(c), ItoaComma(i))
		}
	}
}
