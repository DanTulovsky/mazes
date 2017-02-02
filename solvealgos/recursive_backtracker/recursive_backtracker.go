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

type RecursiveBacktracker struct {
	solvealgos.Common
}

// Step steps into the next cell and returns true if it reach toCell.
func Step(g *grid.Grid, currentCell, toCell *grid.Cell, path *grid.Stack) bool {
	currentCell.SetVisited()
	path.Push(currentCell)

	if currentCell == toCell {
		return true
	}

	for _, nextCell := range currentCell.Links() {
		if !nextCell.Visited() {
			totalStep++
			if Step(g, nextCell, toCell, path) {
				return true
			}
		}

	}
	path.Pop()

	// make sure to count when backtracking
	totalStep++
	currentCell.SetVisited()

	return false
}

func (a *RecursiveBacktracker) Solve(g *grid.Grid, fromCell, toCell *grid.Cell) (*grid.Grid, error) {
	defer solvealgos.TimeTrack(a, time.Now())

	var path = grid.NewStack()

	if r := Step(g, fromCell, toCell, path); !r {
		return nil, fmt.Errorf("failed to find path through maze from %v to %v", fromCell, toCell)
	}

	g.SetPathFromTo(fromCell, toCell, path.List())
	// stats
	a.SetSolvePath(path.List())
	a.SetSolveSteps(totalStep)

	return g, nil
}
