// Package random implements the random walk maze solving algorithm

// Walk around the maze until you find a solution.  Dumb as it gets.
package random

import (
	"mazes/grid"
	"mazes/solvealgos"
	"time"
)

type Random struct {
	solvealgos.Common
}

func (a *Random) Solve(g *grid.Grid, fromCell, toCell *grid.Cell) (*grid.Grid, error) {
	defer solvealgos.TimeTrack(a, time.Now())

	var path = grid.NewPath()
	currentCell := fromCell
	facing := "north" // arbitrary

	for currentCell != toCell {
		path.AddSegement(grid.NewSegment(currentCell, facing))
		currentCell.SetVisited()

		nextCell := currentCell.RandomLink()
		facing = currentCell.GetFacingDirection(nextCell)
		currentCell = nextCell
	}

	// add the last cell
	facing = currentCell.GetFacingDirection(toCell)
	path.AddSegement(grid.NewSegment(toCell, facing))
	g.SetPathFromTo(fromCell, toCell, path.ListCells())

	// stats
	a.SetSolvePath(path)
	a.SetTravelPath(path)
	a.SetSolveSteps(len(path.ListCells()))

	return g, nil
}
