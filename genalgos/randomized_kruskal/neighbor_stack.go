package randomized_kruskal

import (
	"fmt"
	"math/rand"
	"mazes/maze"
)

type neighborPair struct {
	left, right *maze.Cell
}

func (np *neighborPair) String() string {
	return fmt.Sprintf("[%v, %v]", np.left, np.right)
}

// Stack is a stack of *Cell objects
type NeighborStack struct {
	pairs []*neighborPair
}

func NewNeighborStack() *NeighborStack {
	return &NeighborStack{
		pairs: make([]*neighborPair, 0),
	}
}

func (s *NeighborStack) String() string {
	out := ""
	for _, p := range s.pairs {
		out = fmt.Sprintf("%v %v", out, p)
	}
	return out
}

func (s *NeighborStack) Push(c *neighborPair) {
	s.pairs = append(s.pairs, c)

}

func (s *NeighborStack) Pop() (cell *neighborPair) {
	if len(s.pairs) == 0 {
		return nil
	}
	cell, s.pairs = s.pairs[len(s.pairs)-1], s.pairs[:len(s.pairs)-1]
	return cell
}

func (s *NeighborStack) Size() int {
	return len(s.pairs)
}

// Top returns the topmost cell (without popping it off)
func (s *NeighborStack) Top() *neighborPair {
	if len(s.pairs) == 0 {
		return nil
	}
	return s.pairs[len(s.pairs)-1]
}

// List returns the list of cells in the stack
func (s *NeighborStack) List() []*neighborPair {
	return s.pairs
}

// RandomList returns the list of cells in the stack in random order
func (s *NeighborStack) RandomList() []*neighborPair {
	var pairs []*neighborPair
	r := rand.Perm(len(s.pairs))

	for _, i := range r {
		pairs = append(pairs, s.pairs[i])
	}
	return pairs
}

// Shuffle shuffles in stack in place
func (s *NeighborStack) Shuffle() {
	var pairs []*neighborPair
	r := rand.Perm(len(s.pairs))

	for _, i := range r {
		pairs = append(pairs, s.pairs[i])
	}
	s.pairs = pairs
}
