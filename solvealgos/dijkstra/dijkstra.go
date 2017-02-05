// Package dijkstra implements dijkstra's algorithm to find the shortest path
// Note that this algorithm knows the entire layout of the maze and the distances between all cells
package dijkstra

import (
	"math"
	"mazes/grid"
	"mazes/solvealgos"
	"time"
)

type Dijkstra struct {
	solvealgos.Common
}

func (a *Dijkstra) Solve(g *grid.Grid, fromCell, toCell *grid.Cell, delay time.Duration) (*grid.Grid, error) {
	defer solvealgos.TimeTrack(a, time.Now())

	var travelPath = g.TravelPath
	var solvePath = g.SolvePath

	// Get all distances from this cell
	d := fromCell.Distances()

	currentCell := toCell

	for currentCell != d.Root() {
		// animation delay
		time.Sleep(delay)

		smallest := math.MaxInt16
		var next *grid.Cell
		for _, link := range currentCell.Links() {
			dist, _ := d.Get(link)
			if dist < smallest {
				smallest = dist
				next = link
			}
		}
		segment := grid.NewSegment(next, "north") // arbitrary facing
		travelPath.AddSegement(segment)
		solvePath.AddSegement(segment)
		g.SetPathFromTo(fromCell, currentCell, travelPath)
		currentCell = next
	}

	// add toCell to path
	travelPath.ReverseCells()
	segment := grid.NewSegment(toCell, "north") // arbitrary facing
	travelPath.AddSegement(segment)
	solvePath.AddSegement(segment)
	g.SetPathFromTo(fromCell, toCell, travelPath)

	// stats
	a.SetSolvePath(solvePath)
	a.SetTravelPath(travelPath)
	a.SetSolveSteps(solvePath.Length())

	return g, nil
}
