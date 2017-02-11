// Package random implements the random walk maze solving algorithm

// Walk around the maze until you find a solution.  Dumb as it gets.
package random

import (
	"mazes/maze"
	"mazes/solvealgos"
	"time"
)

type Random struct {
	solvealgos.Common
}

func (a *Random) Solve(g *maze.Maze, fromCell, toCell *maze.Cell, delay time.Duration) (*maze.Maze, error) {
	defer solvealgos.TimeTrack(a, time.Now())

	var travelPath = g.TravelPath()
	var solvePath = g.SolvePath()
	currentCell := fromCell
	facing := "north" // arbitrary

	for currentCell != toCell {
		// animation delay
		time.Sleep(delay)

		currentCell.SetVisited()

		segment := maze.NewSegment(currentCell, facing)
		travelPath.AddSegement(segment)
		solvePath.AddSegement(segment)
		g.SetPathFromTo(fromCell, currentCell, travelPath)

		nextCell := currentCell.RandomLink()
		facing = currentCell.GetFacingDirection(nextCell)
		currentCell = nextCell
	}

	// add the last cell
	facing = currentCell.GetFacingDirection(toCell)
	segment := maze.NewSegment(toCell, facing)
	travelPath.AddSegement(segment)
	solvePath.AddSegement(segment)
	g.SetPathFromTo(fromCell, toCell, travelPath)

	// stats
	a.SetSolvePath(solvePath)
	a.SetTravelPath(travelPath)
	a.SetSolveSteps(travelPath.Length())

	return g, nil
}
