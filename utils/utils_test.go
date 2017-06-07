package utils

import "testing"

var affinetransformtests = []struct {
	in         float64
	out        int
	a, b, c, d float64 // [a,b] -> [c,d]
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

var offsettests = []struct {
	in       int
	expected int
}{
	{in: 0, expected: 0},
	{in: 1, expected: 1},
	{in: 2, expected: -1},
	{in: 3, expected: 2},
	{in: 4, expected: -2},
}

func TestDrawOffset(t *testing.T) {
	for _, tt := range offsettests {
		r := DrawOffset(tt.in)
		if r != tt.expected {
			t.Errorf("expected %v for input %v, but got %v", tt.expected, tt.in, r)
		}
	}
}

var oddtests = []struct {
	in       int
	expected bool
}{
	{in: 0, expected: false},
	{in: 1, expected: true},
	{in: 2, expected: false},
	{in: 6, expected: false},
	{in: 3, expected: true},
}

func TestIsOdd(t *testing.T) {
	for _, tt := range oddtests {
		r := isOdd(tt.in)
		if r != tt.expected {
			t.Errorf("expected %v for input %v, but got %v", tt.expected, tt.in, r)
		}
	}
}
