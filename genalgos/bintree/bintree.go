// Package bintree implements the binary tree algorithm for maze generation

// For each cell in the grid, you decide whether to carve a passage north or east.
package bintree

import (
	"fmt"
	"time"

	"github.com/tevino/abool"
	"mazes/genalgos"
	"mazes/maze"
	"mazes/utils"
)

type Bintree struct {
	genalgos.Common
}

// Apply applies the binary tree algorithm to generate the maze.
func (a *Bintree) Apply(m *maze.Maze, delay time.Duration, generating *abool.AtomicBool) error {
	defer genalgos.TimeTrack(m, time.Now())

	for _, currentCell := range m.OrderedCells() {
		if !generating.IsSet() {
			return fmt.Errorf("stop requested")
		}

		time.Sleep(delay) // animation delay
		m.SetGenCurrentLocation(currentCell)

		neighbors := []*maze.Cell{}
		if currentCell.North() != nil {
			neighbors = append(neighbors, currentCell.North())
		}
		if currentCell.East() != nil {
			neighbors = append(neighbors, currentCell.East())
		}

		if len(neighbors) == 0 {
			continue
		}
		index := utils.Random(0, len(neighbors))
		neighbor := neighbors[index]
		if neighbor != nil {
			m.Link(currentCell, neighbor)
		}
	}

	a.Cleanup(m)
	return nil
}
