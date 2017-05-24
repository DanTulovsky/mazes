// Package kruskal implements kruskal's algorithm for maze generation
package kruskal

import (
	"time"

	"mazes/genalgos"
	"mazes/maze"
	"mazes/utils"
)

type state struct {
	maze       *maze.Maze
	neighbors  *NeighborStack
	setForCell map[*maze.Cell]int
	cellsInSet map[int][]*maze.Cell
}

func newState(m *maze.Maze) *state {
	neighbors := NewNeighborStack()
	setForCell := make(map[*maze.Cell]int, 5000)
	cellsInSet := make(map[int][]*maze.Cell, 5000)

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

		// assign random cost to each pair, they pop out for the algorithm from lowest -> highest
		randomCost := utils.Random(0, 100)
		for _, n := range c.Neighbors() {
			neighbors.Push(&neighborPair{left: c, right: n, cost: randomCost})
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

// addCrossing adds a crossing at this cell
func (s *state) addCrossing(c *maze.Cell) bool {
	if len(c.Links()) != 0 {
		return false
	}

	if !s.canMerge(c.East(), c.West()) || !s.canMerge(c.North(), c.South()) {
		return false
	}

	// remove this cell as an option
	s.neighbors.Delete(c)

	// randomly pick the direction of passage
	if utils.Random(0, 2) == 0 {
		s.Merge(c, c.East())
		s.Merge(c.West(), c)
		s.Merge(c.North(), c.South())

	} else {
		s.Merge(c.North(), c)
		s.Merge(c, c.South())
		s.Merge(c.East(), c.West())
	}
	return true
}

type Kruskal struct {
	genalgos.Common
}

// Apply applies the algorithm to the grid.
func (a *Kruskal) Apply(m *maze.Maze, delay time.Duration) error {
	defer genalgos.TimeTrack(m, time.Now())

	s := newState(m)
	cells := maze.CellMapKeys(m.Cells())

	// add crossings (under-passages) as required
	for x := int64(0); x < m.Size(); x++ {
		if !m.Config().AllowWeaving || utils.Random(0, 100) >= int(m.Config().WeavingProbability*100) {
			continue
		}

		cell := cells[utils.Random(0, len(cells))]
		if cell.East() != nil && cell.West() != nil && cell.North() != nil && cell.South() != nil {
			s.addCrossing(cell)
		}
	}

	for s.neighbors.Size() > 0 {
		time.Sleep(delay) // animation delay
		n := s.neighbors.Pop()
		if s.canMerge(n.left, n.right) {
			s.Merge(n.left, n.right)
		}
	}
	a.Cleanup(m)
	return nil
}
