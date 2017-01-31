// Package bintree implements the binary tree algorithm for maze generation

// For each cell in the grid, you decide whether to carve a passage north or east.
package bintree

import (
	"mazes/genalgos"
	"mazes/grid"
	"mazes/utils"
	"time"
)

type Bintree struct {
	genalgos.Common
}

// Apply applies the binary tree algorithm to generate the maze.
func (a *Bintree) Apply(g *grid.Grid) (*grid.Grid, error) {

	defer utils.TimeTrack(time.Now(), "bintree apply")
	for _, cell := range g.Cells() {
		neighbors := []*grid.Cell{}
		if cell.North != nil {
			neighbors = append(neighbors, cell.North)
		}
		if cell.East != nil {
			neighbors = append(neighbors, cell.East)
		}

		if len(neighbors) == 0 {
			continue
		}
		index := utils.Random(0, len(neighbors))
		neighbor := neighbors[index]
		if neighbor != nil {
			cell.Link(neighbor)
		}
	}
	return g, nil
}
