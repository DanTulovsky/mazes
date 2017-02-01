package utils

import "testing"

var affinetransformtests = []struct {
	in         float32
	out        int
	a, b, c, d float32 // [a,b] -> [c,d]
}{
	{in: -5, out: 0, a: 0, b: 10, c: 0, d: 100},
	{in: 50, out: 100, a: 0, b: 10, c: 0, d: 100},
	{in: 1, out: 10, a: 0, b: 10, c: 0, d: 100},
	{in: 10, out: 100, a: 0, b: 10, c: 0, d: 100},
	{in: 3000, out: 239, a: 0, b: 3200, c: 0, d: 255},
	{in: 98, out: 249, a: 0, b: 100, c: 0, d: 255},
}

func TestAffineTransform(t *testing.T) {

	for _, tt := range affinetransformtests {
		r := AffineTransform(tt.in, tt.a, tt.b, tt.c, tt.d)
		if r != tt.out {
			t.Errorf("AffineTransform(%v, %v, %v, %v, %v) => %d, want %d", tt.in, tt.a, tt.b, tt.c, tt.d, r, tt.out)
		}
	}
}
