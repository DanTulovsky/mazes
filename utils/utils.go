package utils

import (
	"log"
	"math/rand"
	"time"
)

// Random returns a random number in [min, max)
func Random(min, max int) int {
	return rand.Intn(max-min) + min
}

// AffineTransform x (in the range [a, b] to a number in [c, d]
func AffineTransform(x, a, b, c, d float64) int {
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

func TimeTrack(start time.Time, name string) {
	elapsed := time.Since(start)
	log.Printf("%s took %s", name, elapsed)
}

func SliceIndex(limit int, predicate func(i int) bool) int {
	for i := 0; i < limit; i++ {
		if predicate(i) {
			return i
		}
	}
	return -1
}

func StrInList(l []string, s string) bool {
	for _, str := range l {
		if str == s {
			return true
		}
	}
	return false
}

func isOdd(n int) bool {
	return !(n%2 == 0)
}

// offset returns the path offset of the given client number n
// 0 -> 0; 1 -> 1; 2 -> -1; 3 -> 2; 4 -> -2
func DrawOffset(n int) int {
	if n == 0 {
		return 0
	}
	if n == 1 {
		return 1
	}

	if isOdd(n) {
		return (n + 1) / 2
	}
	return -(n / 2)
}

// SetBit sets the bit at pos in the integer n.
func SetBit(n int, pos uint) int {
	n |= 1 << pos
	return n
}

// ClearBit clears the bit at pos in n.
func ClearBit(n int, pos uint) int {
	mask := ^(1 << pos)
	n &= mask
	return n
}

// HasBit returns true if bit is set
func HasBit(n int, pos uint) bool {
	val := n & (1 << pos)
	return (val > 0)
}
