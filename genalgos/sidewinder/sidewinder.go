// Package sidewinder implements the widewinder algorithm for maze generation

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
func (a *Sidewinder) Apply(g *maze.Grid, delay time.Duration) (*maze.Grid, error) {
	defer genalgos.TimeTrack(g, time.Now())

	gridWidth, _ := g.Dimensions()

	for _, row := range g.Rows() {
		var run []*maze.Cell

		for x := len(row) - 1; x >= 0; x-- {
			time.Sleep(delay) // animation delay

			cell := row[x]
			g.SetGenCurrentLocation(cell)

			run = append(run, cell)

			// 0 = north, 1 = east
			rand := utils.Random(0, 2)

			if rand == 1 {
				// if possible, open passage east
				if cell.East != nil {
					cell.Link(cell.East)
					continue
				} else if cell.North != nil {
					// close out run, we are at the far right wall
					if x != 0 {
						// something went wrong!
						log.Fatalf("x=%v; expected x=%v (should be at far right)", x, len(row)-1)
					}
					c := g.RandomCellFromList(run)
					if c.North != nil {
						c.Link(c.North)
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
				c := g.RandomCellFromList(run)
				if c.North != nil {
					// open north passage
					c.Link(c.North)
				} else if cell.East != nil {
					// unless you can't, then open the east passage
					cell.Link(cell.East)
				}
				// clear out run
				run = []*maze.Cell{}
			}

		}

	}

	a.Cleanup(g)
	return g, nil
}
