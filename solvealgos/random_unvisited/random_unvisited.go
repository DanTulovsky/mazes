// Package random_unvisited implements the random unvisited walk maze solving algorithm

// Walk around the maze until you find a solution. Prefer unvisited first.
package random_unvisited

import (
	"mazes/grid"
	"mazes/solvealgos"
	"time"
)

type RandomUnvisited struct {
	solvealgos.Common
}

func (a *RandomUnvisited) Solve(g *grid.Grid, fromCell, toCell *grid.Cell) (*grid.Grid, error) {
	defer solvealgos.TimeTrack(a, time.Now())

	var path = grid.NewStack()

	currentCell := fromCell

	for currentCell != toCell {
		path.Push(currentCell)
		currentCell.SetVisited()

		// prefer unvisited first
		nextCell := currentCell.RandomUnvisitedLink()
		if nextCell == nil {
			nextCell = currentCell.RandomLink()
		}

		currentCell = nextCell

	}

	path.Push(toCell)
	g.SetPathFromTo(fromCell, toCell, path.List())
	// stats
	a.SetSolvePath(path.List())
	a.SetTravelPath(path.List())
	a.SetSolveSteps(len(path.List()))

	return g, nil
}
