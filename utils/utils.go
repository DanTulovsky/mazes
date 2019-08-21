package utils

import (
	"log"
	"math/rand"
	"time"

	"fmt"
	"math"

	pb "github.com/DanTulovsky/mazes/proto"

	"github.com/gonum/matrix/mat64"
)

func init() {
	rand.Seed(time.Now().Unix())
}

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

// DirectionInList returns true if the direction s (e.g. 'north') is in the list of directions
func DirectionInList(l []*pb.Direction, s string) bool {
	for _, str := range l {
		if str.Name == s {
			return true
		}
	}
	return false
}

func IsOdd(n int) bool {
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

	if IsOdd(n) {
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

// LocationFromState return the x,y coordinates of the l'th point in an array rows X columns, counting from top left
// one row at a time. Location starts with 1.
func LocationFromState(rows, columns, l int64) (*pb.MazeLocation, error) {
	if l > rows*columns || l < 0 {
		return nil, fmt.Errorf("%v is too large or too small for grid of (columns, rows) [%v, %v]", l, columns, rows)
	}

	r, c := ModDiv(l, columns)
	return &pb.MazeLocation{X: c, Y: r, Z: 0}, nil
}

// StateFromLocation returns the state number given a location
func StateFromLocation(rows, columns int64, l *pb.MazeLocation) (int, error) {
	if l == nil {
		return 0, fmt.Errorf("location is nil...")
	}
	if l.X > columns || l.Y > rows {
		return 0, fmt.Errorf("requested coordinates (%v) are outside the grid (columns, rows) (%v, %v)", l, columns, rows)
	}
	return int(l.X + l.Y*columns), nil
}

func LocsSame(l, m *pb.MazeLocation) bool {
	if l.X == m.X && l.Y == m.Y && l.Z == m.Z {
		return true
	}
	return false
}

// ModDiv returns the result of x/y; integer part and remainder part
// Don't divide by 0...
func ModDiv(x, y int64) (int64, int64) {
	if y == 0 {
		return y, x
	}
	ipart := x / y
	rpart := math.Mod(float64(x), float64(y))

	return ipart, int64(rpart)
}

// WeightedChoice returns the index of randomly chosen element in the vector based on the weight
// e.g. [1, 0, 3, 4, 0] -> will most often return 3 (the index of value 4)
func WeightedChoice(v *mat64.Vector) int {
	totals := []float64{}
	runningTotal := 0.0

	for i := 0; i < v.Len(); i++ {
		w := v.At(i, 0)
		runningTotal += float64(w)
		totals = append(totals, runningTotal)
	}

	rnd := float64(Random(0, 100)) / 100.0 * runningTotal

	for i, total := range totals {
		if rnd < total {
			return i
		}
	}
	return -1
}

// p is the input, t is the time, r is the decay constant
func Decay(p float64, t, r float64) float64 {
	// A = Pe^(rt)
	return p * math.Exp(r*t)
}
