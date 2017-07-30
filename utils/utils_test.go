package utils

import (
	"log"
	pb "mazes/proto"
	"testing"

	"github.com/gonum/matrix/mat64"
)

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

var bittests = []struct {
	in       int
	pos      uint
	expected int
}{
	{in: 0, pos: 0, expected: 1},
	{in: 0, pos: 1, expected: 2},
}

func TestBits(t *testing.T) {
	for _, tt := range bittests {
		r := SetBit(tt.in, tt.pos)
		if r != tt.expected {
			t.Errorf("expected %v for input %v, but got %v", tt.expected, tt.in, r)
		}

		hc := HasBit(r, tt.pos)
		if !hc {
			t.Errorf("expected %v for input %v, but got %v", true, r, false)
		}

		rc := ClearBit(r, tt.pos)
		if rc != tt.in {
			t.Errorf("expected %v for input %v, but got %v", tt.in, r, r, rc)
		}
	}
}

var locationfromstatetests = []struct {
	rows     int64
	columns  int64
	location int64
	expected *pb.MazeLocation
	wantErr  bool
}{
	{rows: 10, columns: 10, location: 0, expected: &pb.MazeLocation{0, 0, 0}, wantErr: false},
	{rows: 3, columns: 3, location: 5, expected: &pb.MazeLocation{2, 1, 0}, wantErr: false},
	{rows: 3, columns: 3, location: 60, expected: nil, wantErr: true},
	{rows: 3, columns: 3, location: -1, expected: nil, wantErr: true},
	{rows: 3, columns: 2, location: 4, expected: &pb.MazeLocation{0, 2, 0}, wantErr: false},
	{rows: 2, columns: 2, location: 2, expected: &pb.MazeLocation{0, 1, 0}, wantErr: false},
}

func TestLocationFromState(t *testing.T) {
	for _, tt := range locationfromstatetests {
		l, err := LocationFromState(tt.rows, tt.columns, tt.location)
		if err != nil {
			if !tt.wantErr {
				t.Fatalf("error getting location (%v): %v", tt.location, err)
			} else {
				continue // skip the rest of the test
			}
		}

		if tt.expected.X != l.X || tt.expected.Y != l.Y || tt.expected.Z != l.Z {
			t.Errorf("expected: %v; received: %v (rows=%v, columns=%v)", tt.expected, l, tt.rows, tt.columns)
		}
	}
}

var statefromlocationtests = []struct {
	l        *pb.MazeLocation
	rows     int64
	columns  int64
	expected int
	wantErr  bool
}{
	{l: &pb.MazeLocation{0, 0, 0}, rows: 10, columns: 10, expected: 0, wantErr: false},
	{l: &pb.MazeLocation{23, 0, 0}, rows: 10, columns: 10, expected: 0, wantErr: true},
	{l: &pb.MazeLocation{1, 0, 0}, rows: 10, columns: 10, expected: 1, wantErr: false},
	{l: &pb.MazeLocation{2, 1, 0}, rows: 3, columns: 4, expected: 6, wantErr: false},
	{l: &pb.MazeLocation{0, 2, 0}, rows: 4, columns: 3, expected: 6, wantErr: false},
}

func TestStateFromLocation(t *testing.T) {
	for _, tt := range statefromlocationtests {
		state, err := StateFromLocation(tt.rows, tt.columns, tt.l)
		if err != nil {
			if !tt.wantErr {
				t.Fatalf("failed to find state from l (%v)", tt.l, err)
			} else {
				continue
			}
		}

		if state != tt.expected {
			t.Errorf("expected: %v; received: %v; location: %v", tt.expected, state, tt.l)
		}

	}
}

var locsametests = []struct {
	l        *pb.MazeLocation
	m        *pb.MazeLocation
	expected bool
}{
	{l: &pb.MazeLocation{0, 0, 0}, m: &pb.MazeLocation{0, 0, 0}, expected: true},
	{l: &pb.MazeLocation{3, 2, 6}, m: &pb.MazeLocation{3, 2, 6}, expected: true},
	{l: &pb.MazeLocation{0, 0, 0}, m: &pb.MazeLocation{0, 1, 0}, expected: false},
	{l: &pb.MazeLocation{0, 0, 0}, m: &pb.MazeLocation{3, 0, 0}, expected: false},
}

func TestLocsSame(t *testing.T) {
	for _, tt := range locsametests {
		r := LocsSame(tt.l, tt.m)
		if r != tt.expected {
			t.Errorf("expected: %v, received: %v; l: %v; m: %v", tt.expected, r, tt.l, tt.m)
		}

	}
}

var moddivtests = []struct {
	x            int64
	y            int64
	expWhole     int64
	expRemainder int64
}{
	{x: 7, y: 4, expWhole: 1, expRemainder: 3},
	{x: 0, y: 0, expWhole: 0, expRemainder: 0},
	{x: 3, y: 0, expWhole: 0, expRemainder: 3},
}

func TestModDiv(t *testing.T) {
	for _, tt := range moddivtests {
		w, r := ModDiv(tt.x, tt.y)

		if w != tt.expWhole || r != tt.expRemainder {
			t.Errorf("expected: %vr%v; got %vr%v", tt.expWhole, tt.expRemainder, w, r)
		}

	}
}

var weightedchoicetest = []struct {
	v *mat64.Vector
}{
	{v: mat64.NewVector(5, []float64{.1, 0, 0, 0, 0})},
	{v: mat64.NewVector(5, []float64{0, 10, 0, 0, 0})},
	{v: mat64.NewVector(5, []float64{.2, 10, .2, 10, .2})},
	{v: mat64.NewVector(5, []float64{.2, .3, .2, .2, .2})},
}

func TestWeightedChoice(t *testing.T) {
	for _, tt := range weightedchoicetest {
		c := WeightedChoice(tt.v)
		log.Printf("index: %v; value: %v; vector: %v", c, tt.v.At(c, 0), tt.v)

	}
}

var weightedchoicebench = []struct {
	v *mat64.Vector
}{
	{v: mat64.NewVector(4, []float64{0, 0, 0.1, 0})},
}

func BenchmarkWeightedChoice(t *testing.B) {
	for _, tt := range weightedchoicebench {
		c := WeightedChoice(tt.v)
		if c != 2 {
			t.Errorf("picked element with probability 0: %v", c)
		}

	}
}
