// Package recursive_backtracker implements the recursive backtracker  maze solving algorithm

package recursive_backtracker

import (
	"fmt"
	"mazes/grid"
	"mazes/solvealgos"
	"time"
)

// total steps take during walk of maze
var totalStep int = 1
var travelPath *grid.Path = grid.NewPath()
var facing string = "north"

type RecursiveBacktracker struct {
	solvealgos.Common
}

// Step steps into the next cell and returns true if it reach toCell.
func Step(g *grid.Grid, currentCell, toCell *grid.Cell, path *grid.Path) bool {
	var nextCell *grid.Cell
	currentCell.SetVisited()

	path.AddSegement(grid.NewSegment(currentCell, facing))
	travelPath.AddSegement(grid.NewSegment(currentCell, facing))

	if currentCell == toCell {
		return true
	}

	for _, nextCell = range currentCell.Links() {
		if !nextCell.Visited() {
			facing = currentCell.GetFacingDirection(nextCell)
			totalStep++
			if Step(g, nextCell, toCell, path) {
				return true
			}
		}

	}
	path.DelSegement()

	// make sure to count when backtracking
	totalStep++
	currentCell.SetVisited()

	facing = nextCell.GetFacingDirection(currentCell)
	travelPath.AddSegement(grid.NewSegment(currentCell, facing))

	return false
}

func (a *RecursiveBacktracker) Solve(g *grid.Grid, fromCell, toCell *grid.Cell) (*grid.Grid, error) {
	defer solvealgos.TimeTrack(a, time.Now())

	totalStep = 1
	var path = grid.NewPath()

	if r := Step(g, fromCell, toCell, path); !r {
		return nil, fmt.Errorf("failed to find path through maze from %v to %v", fromCell, toCell)
	}

	g.SetPathFromTo(fromCell, toCell, path.ListCells())
	// stats
	a.SetSolvePath(path)
	a.SetSolveSteps(totalStep)
	a.SetTravelPath(travelPath)

	return g, nil
}
