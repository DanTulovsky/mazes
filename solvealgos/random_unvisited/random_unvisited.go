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

func (a *RandomUnvisited) Solve(g *grid.Grid, fromCell, toCell *grid.Cell, delay time.Duration) (*grid.Grid, error) {
	defer solvealgos.TimeTrack(a, time.Now())

	var travelPath = g.TravelPath
	var solvePath = g.SolvePath
	currentCell := fromCell
	facing := "north"

	for currentCell != toCell {
		// animation delay
		time.Sleep(delay)

		currentCell.SetVisited()

		segment := grid.NewSegment(currentCell, facing)
		travelPath.AddSegement(segment)
		solvePath.AddSegement(segment)
		g.SetPathFromTo(fromCell, currentCell, travelPath.ListCells())

		// prefer unvisited first
		nextCell := currentCell.RandomUnvisitedLink()

		if nextCell == nil {
			nextCell = currentCell.RandomLink()
		}

		facing = currentCell.GetFacingDirection(nextCell)
		currentCell = nextCell

	}

	// last cell
	facing = currentCell.GetFacingDirection(toCell)
	segment := grid.NewSegment(currentCell, facing)
	travelPath.AddSegement(segment)
	travelPath.AddSegement(segment)
	g.SetPathFromTo(fromCell, toCell, solvePath.ListCells())

	// stats
	a.SetSolvePath(solvePath)
	a.SetTravelPath(travelPath)
	a.SetSolveSteps(len(solvePath.ListCells()))

	return g, nil
}