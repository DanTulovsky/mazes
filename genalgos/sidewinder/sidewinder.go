// Package sidewinder implements the sidewinder algorithm for maze generation

// Binary Tree chooses between north and east at every cell; Sidewinder, on the other hand, tries to group adjacent
// cells together before carving a passage north from one of them.
package sidewinder

import (
	"log"
	"mazes/genalgos"
	"mazes/maze"
	"mazes/utils"
	"time"
)

type Sidewinder struct {
	genalgos.Common
}

// Apply applies the algorithm to the grid.
func (a *Sidewinder) Apply(m *maze.Maze, delay time.Duration) error {
	defer genalgos.TimeTrack(m, time.Now())

	gridWidth, _ := m.Dimensions()

	for _, row := range m.Rows() {
		var run []*maze.Cell

		for x := len(row) - 1; x >= 0; x-- {
			time.Sleep(delay) // animation delay

			cell := row[x]
			m.SetGenCurrentLocation(cell)

			run = append(run, cell)

			// 0 = north, 1 = east
			rand := utils.Random(0, 2)

			if rand == 1 {
				// if possible, open passage east
				if cell.East() != nil {
					m.Link(cell, cell.East())
					continue
				} else if cell.North != nil {
					// close out run, we are at the far right wall
					if x != 0 {
						// something went wrong!
						log.Fatalf("x=%v; expected x=%v (should be at far right)", x, len(row)-1)
					}
					c := m.RandomCellFromList(run)
					if c.North() != nil {
						m.Link(c, c.North())
					}
					// clear out run
					run = []*maze.Cell{} // not strictly necessary

					continue
				}

				l := cell.Location()
				if l.X != gridWidth-1 || l.Y != 0 {
					log.Fatalf("in cell %v, which is not top-right cell", cell)
				}
				run = []*maze.Cell{}
				continue // should only happen at top right cell
			}

			if rand == 0 {
				// close out run, pick random cell
				c := m.RandomCellFromList(run)
				if c.North() != nil {
					// open north passage
					m.Link(c, c.North())
				} else if cell.East() != nil {
					// unless you can't, then open the east passage
					m.Link(cell, cell.East())
				}
				// clear out run
				run = []*maze.Cell{}
			}

		}

	}

	a.Cleanup(m)
	return nil
}
