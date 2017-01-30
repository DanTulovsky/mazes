package utils

import (
	"log"
	"math/rand"
)

func Random(min, max int) int {
	return rand.Intn(max-min) + min
}

// AffineTransform x (in the range [a, b] to a number in [c, d]
func AffineTransform(x, a, b, c, d float32) int {
	// log.Printf("in: %v [%v, %v] -> [%v, %v]", x, a, b, c, d)
	if x < a {
		log.Print("invalid input into AffineTransform, returning min.")
		log.Printf("AffineTransform -> in: %v [%v, %v] -> [%v, %v]", x, a, b, c, d)
		return int(c)
	}
	if x > b {
		log.Print("invalid input into AffineTransform, returning max.")
		log.Printf("AffineTransform -> in: %v [%v, %v] -> [%v, %v]", x, a, b, c, d)
		return int(d)
	}
	return int((x-a)*((d-c)/(b-a)) + c)
}
