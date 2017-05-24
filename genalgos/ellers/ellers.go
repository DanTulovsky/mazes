// Package ellers implements eller's algorithm for maze generation
package ellers

// Weaving does not work for this algorithm because it never references cell.neighbors()

import (
	"math/rand"
	"time"

	"mazes/genalgos"
	"mazes/maze"
	"mazes/utils"
)

type state struct {
	maze       *maze.Maze
	setForCell map[int64]int64
	cellsInSet map[int64][]*maze.Cell
	nextSet    int64
}

func newState(m *maze.Maze, nextSet int64) *state {

	// column of cell -> set; algorithm works at row level
	setForCell := make(map[int64]int64, 5000)
	cellsInSet := make(map[int64][]*maze.Cell, 5000)

	s := &state{
		maze:       m,
		setForCell: setForCell,
		cellsInSet: cellsInSet,
		nextSet:    nextSet,
	}

	return s
}

func (s *state) record(set int64, c *maze.Cell) {
	s.setForCell[c.Location().X] = set

	if _, ok := s.cellsInSet[set]; !ok {
		s.cellsInSet[set] = []*maze.Cell{}
	}
	s.cellsInSet[set] = append(s.cellsInSet[set], c)
}

// setFor returns the set for a cell. If the cell is not in a set, it assigns it the next one.
func (s *state) setFor(c *maze.Cell) int64 {
	if _, ok := s.setForCell[c.Location().X]; !ok {
		// assign to next set
		s.record(s.nextSet, c)
		s.nextSet++
	}
	return s.setForCell[c.Location().X]
}

// Merge moves all cells in loser set into the winner set
func (s *state) Merge(winner, loser int64) {
	for _, c := range s.cellsInSet[loser] {
		s.setForCell[c.Location().X] = winner
		s.cellsInSet[winner] = append(s.cellsInSet[winner], c)
	}
	delete(s.cellsInSet, loser)
}

// Next returns a new row state counting off from where the previous one left off
func (s *state) Next() *state {
	return newState(s.maze, s.nextSet)
}

func (s *state) CellsInSet() map[int64][]*maze.Cell {
	return s.cellsInSet
}

type Ellers struct {
	genalgos.Common
}

func shuffleCells(cells []*maze.Cell) {
	for i := range cells {
		j := rand.Intn(i + 1)
		cells[i], cells[j] = cells[j], cells[i]
	}
}

// Apply applies the binary tree algorithm to generate the maze.
func (a *Ellers) Apply(m *maze.Maze, delay time.Duration) error {

	defer genalgos.TimeTrack(m, time.Now())

	// initial state
	s := newState(m, 0)

	for _, row := range m.Rows() {
		// pick which cells to merge
		for _, c := range row {
			time.Sleep(delay) // animation delay

			if c.West() == nil {
				continue
			}
			set := s.setFor(c)
			prior_set := s.setFor(c.West())

			var shouldLink bool
			// link if in different sets and if it's last row, or randomly
			if set != prior_set && (c.North() == nil || utils.Random(0, 2) == 0) {
				shouldLink = true
			}

			if shouldLink {
				m.Link(c, c.West())
				s.Merge(prior_set, set)
			}
		}

		// pick which cells to link north
		if row[0].North() != nil {
			// only do this if not the last row
			nextRow := s.Next()

			for _, cells := range s.CellsInSet() {
				time.Sleep(delay) // animation delay

				// shuffle list of cells
				shuffleCells(cells)
				for i, c := range cells {
					// we require at least one cell to link north
					// so pick index 0, the other cells have a 1/3 chances
					// of being linked
					if i == 0 || utils.Random(0, 3) == 0 {
						m.Link(c, c.North())
						nextRow.record(s.setFor(c), c.North())
					}
				}
			}
			// move on to next row
			s = nextRow
		}
	}

	a.Cleanup(m)
	return nil
}
