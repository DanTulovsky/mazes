// Package aldous_broder implements the Adlous-Broder algorithm

// Start anywhere in the grid you want, and choose a random neighbor. Move to that neighbor, and if it hasn’t
// previously been visited, link it to the prior cell. Repeat until every cell has been visited.
package aldous_broder

import (
	"mazes/genalgos"
	"mazes/grid"
	"mazes/utils"
	"time"
)

type AldousBroder struct {
	genalgos.Common
}

// Apply applies the adlous-broder algorithm to generate the maze.
func (a *AldousBroder) Apply(g *grid.Grid) (*grid.Grid, error) {

	defer utils.TimeTrack(time.Now(), "aldos-broder apply")
	var visitedCells int
	currentCell := g.RandomCell()

	for visitedCells < len(g.Cells()) {
		neighbors := currentCell.Neighbors()

		randomNeighbor := g.RandomCellFromList(neighbors)
		if !randomNeighbor.Visited() {
			visitedCells++
			currentCell.Link(randomNeighbor)
			randomNeighbor.SetVisited()
		}
		currentCell = randomNeighbor
	}
	return g, nil

}
