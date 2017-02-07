// Package random_unvisited implements the random unvisited walk maze solving algorithm

// Walk around the maze until you find a solution. Prefer unvisited first.
package random_unvisited

import (
	"mazes/maze"
	"mazes/solvealgos"
	"time"
)

type RandomUnvisited struct {
	solvealgos.Common
}

func (a *RandomUnvisited) Solve(g *maze.Maze, fromCell, toCell *maze.Cell, delay time.Duration) (*maze.Maze, error) {
	defer solvealgos.TimeTrack(a, time.Now())

	var travelPath = g.TravelPath
	var solvePath = g.SolvePath
	currentCell := fromCell
	facing := "north"

	for currentCell != toCell {
		// animation delay
		time.Sleep(delay)

		currentCell.SetVisited()

		segment := maze.NewSegment(currentCell, facing)
		travelPath.AddSegement(segment)
		solvePath.AddSegement(segment)
		g.SetPathFromTo(fromCell, currentCell, travelPath)

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
	segment := maze.NewSegment(currentCell, facing)
	travelPath.AddSegement(segment)
	travelPath.AddSegement(segment)
	g.SetPathFromTo(fromCell, toCell, solvePath)

	// stats
	a.SetSolvePath(solvePath)
	a.SetTravelPath(travelPath)
	a.SetSolveSteps(solvePath.Length())

	return g, nil
}
