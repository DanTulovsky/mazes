// Package recursive_backtracker implements the recursive backtracker  maze solving algorithm

// TODO(dant): Fix me!
package recursive_backtracker

import (
	"fmt"
	"mazes/grid"
	"mazes/solvealgos"
	"time"
)

var travelPath *grid.Path
var facing string = "north"
var startCell *grid.Cell

type RecursiveBacktracker struct {
	solvealgos.Common
}

// Step steps into the next cell and returns true if it reach toCell.
func Step(g *grid.Grid, currentCell, toCell *grid.Cell, path *grid.Path, delay time.Duration) bool {
	// animation delay
	time.Sleep(delay)

	var nextCell *grid.Cell
	currentCell.SetVisited()

	segment := grid.NewSegment(currentCell, facing)
	path.AddSegement(segment)
	travelPath.AddSegement(segment)
	g.SetPathFromTo(startCell, currentCell, travelPath.ListCells())

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

		segmentReturn := grid.NewSegment(currentCell, facing)
		travelPath.AddSegement(segmentReturn)
		currentCell.SetVisited()
		g.SetPathFromTo(startCell, currentCell, travelPath.ListCells())

	}
	path.DelSegement()
	time.Sleep(delay)

	return false
}

func (a *RecursiveBacktracker) Solve(g *grid.Grid, fromCell, toCell *grid.Cell, delay time.Duration) (*grid.Grid, error) {
	defer solvealgos.TimeTrack(a, time.Now())

	var path = g.SolvePath
	travelPath = g.TravelPath
	startCell = fromCell

	if r := Step(g, fromCell, toCell, path, delay); !r {
		return nil, fmt.Errorf("failed to find path through maze from %v to %v", fromCell, toCell)
	}

	g.SetPathFromTo(fromCell, toCell, path.ListCells())

	// stats
	a.SetSolvePath(path)
	a.SetSolveSteps(len(travelPath.ListCells()))
	a.SetTravelPath(travelPath)

	return g, nil
}
