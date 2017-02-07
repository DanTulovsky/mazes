// Package recursive_backtracker implements the recursive backtracker  maze solving algorithm

// TODO(dant): Fix me!
package recursive_backtracker

import (
	"fmt"
	"mazes/maze"
	"mazes/solvealgos"
	"time"
)

var travelPath *maze.Path
var facing string = "north"
var startCell *maze.Cell

type RecursiveBacktracker struct {
	solvealgos.Common
}

// Step steps into the next cell and returns true if it reach toCell.
func Step(g *maze.Maze, currentCell, toCell *maze.Cell, path *maze.Path, delay time.Duration) bool {
	// animation delay
	time.Sleep(delay)

	var nextCell *maze.Cell
	currentCell.SetVisited()

	segment := maze.NewSegment(currentCell, facing)
	path.AddSegement(segment)
	travelPath.AddSegement(segment)
	g.SetPathFromTo(startCell, currentCell, travelPath)

	if currentCell == toCell {
		return true
	}

	for _, nextCell = range currentCell.Links() {
		if !nextCell.Visited() {
			facing = currentCell.GetFacingDirection(nextCell)
			segment.UpdateFacingDirection(facing)
			if Step(g, nextCell, toCell, path, delay) {
				return true
			}
		}

		facing = nextCell.GetFacingDirection(currentCell)

		// don't add the same segment if it's already the last one
		if travelPath.LastSegment().Cell() == currentCell {
			continue
		}

		segmentReturn := maze.NewSegment(currentCell, facing)
		travelPath.AddSegement(segmentReturn)
		currentCell.SetVisited()
		g.SetPathFromTo(startCell, currentCell, travelPath)

	}
	path.DelSegement()
	time.Sleep(delay)

	return false
}

func (a *RecursiveBacktracker) Solve(g *maze.Maze, fromCell, toCell *maze.Cell, delay time.Duration) (*maze.Maze, error) {
	defer solvealgos.TimeTrack(a, time.Now())

	var path = g.SolvePath
	travelPath = g.TravelPath
	startCell = fromCell

	// DFS traversal of the grid
	if r := Step(g, fromCell, toCell, path, delay); !r {
		return nil, fmt.Errorf("failed to find path through maze from %v to %v", fromCell, toCell)
	}

	g.SetPathFromTo(fromCell, toCell, path)

	// stats
	a.SetSolvePath(path)
	a.SetSolveSteps(travelPath.Length())
	a.SetTravelPath(travelPath)

	return g, nil
}
