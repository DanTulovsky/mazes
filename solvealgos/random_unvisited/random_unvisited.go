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

	var path = grid.NewPath()
	var facing = "north"

	currentCell := fromCell

	for currentCell != toCell {
		facing = currentCell.GetFacingDirection(toCell)
		path.AddSegement(grid.NewSegment(toCell, facing))
		currentCell.SetVisited()

		// prefer unvisited first
		nextCell := currentCell.RandomUnvisitedLink()
		if nextCell == nil {
			nextCell = currentCell.RandomLink()
		}

		currentCell = nextCell

	}

	facing = currentCell.GetFacingDirection(toCell)
	path.AddSegement(grid.NewSegment(toCell, facing))

	g.SetPathFromTo(fromCell, toCell, path.ListCells())
	// stats
	a.SetSolvePath(path)
	a.SetTravelPath(path)
	a.SetSolveSteps(len(path.ListCells()))

	return g, nil
}
