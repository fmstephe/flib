package fstrconv

import (
	"bytes"
)

func ItoaComma(i int64) string {
	return ItoaDelim(i, ',')
}

func ItoaDelim(i int64, delim byte) string {
	if i == 0 {
		return "0"
	}
	var b bytes.Buffer
	neg := i < 0
	if neg {
		i = -i
	}
	for cnt := 0; i != 0; cnt++ {
		if cnt == 3 {
			b.WriteByte(delim)
			cnt = 0
		}
		r := i % 10
		i = i / 10
		b.WriteByte(byte(r) + 48)
	}
	if neg {
		b.WriteByte('-')
	}
	return reverse(b.String())
}

func reverse(s string) string {
	// With thanks to Russ Cox
	n := 0
	rune := make([]rune, len(s))
	for _, r := range s {
		rune[n] = r
		n++
	}
	rune = rune[0:n]
	// Reverse
	for i := 0; i < n/2; i++ {
		rune[i], rune[n-1-i] = rune[n-1-i], rune[i]
	}
	// Convert back to UTF-8.
	return string(rune)
}
