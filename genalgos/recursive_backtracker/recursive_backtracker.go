// Package recursive_backtracker implements recursive backtracker maze egenration algorithm

// The Recursive Backtracker algorithm works very much like Hunt-and-Kill, relying on a
// constrained random walk to weave its rivery way across our grid. The difference is in how it
// recovers from dead ends; instead of hunting for another viable cell, it backtracks, retracing
// its steps until it finds a cell that has an unvisited neighbor.
package recursive_backtracker

import (
	"mazes/genalgos"
	"mazes/grid"
	"time"
)

type RecursiveBacktracker struct {
	genalgos.Common
}

// Apply applies the recursive backtracker algorithm to generate the maze.
func (a *RecursiveBacktracker) Apply(g *grid.Grid, delay time.Duration) (*grid.Grid, error) {
	defer genalgos.TimeTrack(g, time.Now())

	cells := grid.NewStack()
	currentCell := g.RandomCell()
	cells.Push(currentCell)

	for currentCell != nil {
		time.Sleep(delay) // animation delay
		currentCell = cells.Top()
		currentCell.SetVisited()
		g.SetGenCurrentLocation(currentCell)

		neighbors := currentCell.Neighbors()

		randomNeighbor := genalgos.RandomUnvisitedCellFromList(neighbors)

		if randomNeighbor == nil {
			// no more unvisited neighbors, go back
			cells.Pop()
			currentCell = cells.Top()
			continue
		}

		currentCell.Link(randomNeighbor)
		cells.Push(randomNeighbor)
	}

	a.Cleanup(g)
	return g, nil
}
