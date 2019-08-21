// Package recursive_backtracker implements recursive backtracker maze egenration algorithm

// The Recursive Backtracker algorithm works very much like Hunt-and-Kill, relying on a
// constrained random walk to weave its rivery way across our grid. The difference is in how it
// recovers from dead ends; instead of hunting for another viable cell, it backtracks, retracing
// its steps until it finds a cell that has an unvisited neighbor.
package recursive_backtracker

import (
	"fmt"
	"time"

	"github.com/tevino/abool"
	"github.com/DanTulovsky/mazes/genalgos"
	"github.com/DanTulovsky/mazes/maze"
)

type RecursiveBacktracker struct {
	genalgos.Common
}

// Apply applies the recursive backtracker algorithm to generate the maze.
func (a *RecursiveBacktracker) Apply(m *maze.Maze, delay time.Duration, generating *abool.AtomicBool) error {
	defer genalgos.TimeTrack(m, time.Now())

	cells := maze.NewStack()
	currentCell := m.RandomCell()
	cells.Push(currentCell)

	for currentCell != nil {
		if !generating.IsSet() {
			return fmt.Errorf("stop requested")
		}

		time.Sleep(delay) // animation delay

		currentCell = cells.Top()
		currentCell.SetVisited(maze.VisitedGenerator)
		m.SetGenCurrentLocation(currentCell)

		neighbors := currentCell.Neighbors()

		randomNeighbor := genalgos.RandomUnvisitedCellFromList(neighbors)

		if randomNeighbor == nil {
			// no more unvisited neighbors, go back
			cells.Pop()
			currentCell = cells.Top()
			continue
		}

		m.Link(currentCell, randomNeighbor)
		cells.Push(randomNeighbor)
	}

	a.Cleanup(m)
	return nil
}
