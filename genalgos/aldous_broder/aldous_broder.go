// Package aldous_broder implements the Adlous-Broder algorithm

// Start anywhere in the grid you want, and choose a random neighbor. Move to that neighbor, and if it hasn’t
// previously been visited, link it to the prior cell. Repeat until every cell has been visited.
package aldous_broder

import (
	"mazes/genalgos"
	"mazes/maze"
	"time"
)

type AldousBroder struct {
	genalgos.Common
}

// Apply applies the adlous-broder algorithm to generate the maze.
func (a *AldousBroder) Apply(m *maze.Maze, delay time.Duration) error {
	defer genalgos.TimeTrack(m, time.Now())

	var visitedCells int
	currentCell := m.RandomCell()
	currentCell.SetVisited()
	visitedCells++

	for visitedCells < len(m.Cells()) {
		time.Sleep(delay) // animation delay
		m.SetGenCurrentLocation(currentCell)

		neighbors := currentCell.Neighbors()

		randomNeighbor := m.RandomCellFromList(neighbors)
		if !randomNeighbor.Visited() {
			visitedCells++
			m.Link(currentCell, randomNeighbor)
			randomNeighbor.SetVisited()
		}
		currentCell = randomNeighbor
	}

	a.Cleanup(m)
	return nil
}
