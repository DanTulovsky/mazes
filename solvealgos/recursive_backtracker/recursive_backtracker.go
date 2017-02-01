package recursive_backtracker

import (
	"log"
	"mazes/grid"
	"mazes/solvealgos"
	"time"
)

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
			if Step(g, nextCell, toCell, path) {
				return true
			}
		}

	}
	path.Pop()
	return false
}

func (a *RecursiveBacktracker) Solve(g *grid.Grid, fromCell, toCell *grid.Cell) (*grid.Grid, error) {
	defer solvealgos.TimeTrack(a, time.Now())

	var path = grid.NewStack()

	if r := Step(g, fromCell, toCell, path); !r {
		log.Printf("failed to find path through maze from %v to %v", fromCell, toCell)
	}

	g.SetPathFromTo(fromCell, toCell, path.List())
	// stats
	a.SetSolvePath(path.List())

	return g, nil
}
