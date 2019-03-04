package kruskal

import (
	"container/heap"
	"fmt"
	"gogs.wetsnow.com/dant/mazes/maze"
)

type neighborPair struct {
	left, right *maze.Cell
	cost        int
}

func (np *neighborPair) String() string {
	return fmt.Sprintf("[%v, %v]", np.left, np.right)
}

// Weight returns the cost of moving between the cells
func (np *neighborPair) Weight() int {
	return np.cost
}

// NeighborStack is a stack of *neighborPair objects
type NeighborStack struct {
	// pairs []*neighborPair
	pairs *NeighborPairPriorityQueue
}

func NewNeighborStack() *NeighborStack {
	// make a priority queue to hold the pairs
	queue := make(NeighborPairPriorityQueue, 0)
	heap.Init(&queue)

	return &NeighborStack{
		pairs: &queue,
	}
}

func (s *NeighborStack) String() string {
	out := "queue"
	//for _, p := range s.pairs {
	//	out = fmt.Sprintf("%v %v", out, p)
	//}
	for x := 0; x < s.pairs.Len(); x++ {
		out = fmt.Sprintf("%v %v", out, (*s.pairs)[x])
	}
	return out
}

func (s *NeighborStack) Push(p *neighborPair) {
	heap.Push(s.pairs, p)

}

func (s *NeighborStack) Pop() *neighborPair {
	if s.pairs.Len() == 0 {
		return nil
	}
	return heap.Pop(s.pairs).(*neighborPair)
}

// Delete removes any pairs that include cell c.
func (s *NeighborStack) Delete(c *maze.Cell) {

	for x := 0; x < s.pairs.Len(); x++ {
		if (*s.pairs)[x].left == c || (*s.pairs)[x].right == c {
			heap.Remove(s.pairs, x)
		}
	}
}

func (s *NeighborStack) Size() int {
	return s.pairs.Len()
}
