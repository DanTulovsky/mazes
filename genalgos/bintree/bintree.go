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
func (a *Bintree) Apply(g *grid.Grid, delay time.Duration) (*grid.Grid, error) {
	defer genalgos.TimeTrack(g, time.Now())

	for _, currentCell := range g.Cells() {
		time.Sleep(delay) // animation delay
		g.SetGenCurrentLocation(currentCell)

		neighbors := []*grid.Cell{}
		if currentCell.North != nil {
			neighbors = append(neighbors, currentCell.North)
		}
		if currentCell.East != nil {
			neighbors = append(neighbors, currentCell.East)
		}

		if len(neighbors) == 0 {
			continue
		}
		index := utils.Random(0, len(neighbors))
		neighbor := neighbors[index]
		if neighbor != nil {
			currentCell.Link(neighbor)
		}
	}
	return g, nil
}
