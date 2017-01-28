package utils

import "math/rand"

func Random(min, max int) int {
	return rand.Intn(max-min) + min
}

// AffineTransform x (in the range [a, b] to a number in [c, d]
func AffineTransform(x, a, b, c, d int) int {
	return (x-a)*((d-c)/(b-a)) + c
}
