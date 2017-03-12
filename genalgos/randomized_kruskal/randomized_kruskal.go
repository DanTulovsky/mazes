// Package kruskal implements kruskal's algorithm for maze generation
package randomized_kruskal

import (
	"mazes/genalgos"
	"mazes/maze"
	"time"
)

type state struct {
	maze       *maze.Maze
	neighbors  *NeighborStack
	setForCell map[*maze.Cell]int
	cellsInSet map[int][]*maze.Cell
}

func newState(m *maze.Maze) *state {
	neighbors := NewNeighborStack()
	setForCell := make(map[*maze.Cell]int, 0)
	cellsInSet := make(map[int][]*maze.Cell, 0)

	s := &state{
		maze:       m,
		neighbors:  neighbors,
		setForCell: setForCell,
		cellsInSet: cellsInSet,
	}

	for c := range m.Cells() {
		set := len(setForCell)

		// add cell into its own set
		setForCell[c] = set
		cellsInSet[set] = []*maze.Cell{c}

		// This does duplicate work by checking all neighbors of all cells,
		// but this is required due to how drawing is implemented
		for _, n := range c.Neighbors() {
			neighbors.Push(&neighborPair{c, n})
		}

	}

	return s
}

// canMerge returns true if the two cells are in different sets, and, so, can be merged
func (s *state) canMerge(left, right *maze.Cell) bool {
	return s.setForCell[left] != s.setForCell[right]
}

// Merge combines two sets of cells together
func (s *state) Merge(left, right *maze.Cell) {
	s.maze.Link(left, right)

	winner := s.setForCell[left] // this set remains
	loser := s.setForCell[right] // this is is deleted

	losers := s.cellsInSet[loser]

	// re-assign losing set cells to the winner set
	for _, c := range losers {
		s.cellsInSet[winner] = append(s.cellsInSet[winner], c)
		s.setForCell[c] = winner
	}

	// delete the losing set
	delete(s.cellsInSet, loser)
}

type RandomizedKruskal struct {
	genalgos.Common
}

// Apply applies the algorithm to the grid.
func (a *RandomizedKruskal) Apply(m *maze.Maze, delay time.Duration) (*maze.Maze, error) {
	defer genalgos.TimeTrack(m, time.Now())

	s := newState(m)
	s.neighbors.Shuffle()

	for s.neighbors.Size() > 0 {
		time.Sleep(delay) // animation delay
		n := s.neighbors.Pop()
		if s.canMerge(n.left, n.right) {
			s.Merge(n.left, n.right)
		}
	}
	a.Cleanup(m)
	return m, nil
}
