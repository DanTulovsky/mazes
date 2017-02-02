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

	var path = grid.NewStack()

	currentCell := fromCell

	for currentCell != toCell {
		path.Push(currentCell)
		currentCell.SetVisited()

		nextCell := currentCell.RandomLink()
		currentCell = nextCell
	}

	path.Push(toCell)
	g.SetPathFromTo(fromCell, toCell, path.List())
	// stats
	a.SetSolvePath(path.List())
	a.SetSolveSteps(len(path.List()))

	return g, nil
}
