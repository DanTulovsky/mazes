// Package dijkstra implements dijkstra's algorithm to find the shortest path
// Note that this algorithm knows the entire layout of the maze and the distances between all cells
package dijkstra

import (
	"math"
	"mazes/maze"
	"mazes/solvealgos"
	"time"
)

type Dijkstra struct {
	solvealgos.Common
}

func (a *Dijkstra) Solve(g *maze.Maze, fromCell, toCell *maze.Cell, delay time.Duration) (*maze.Maze, error) {
	defer solvealgos.TimeTrack(a, time.Now())

	// swap these for proper drawing colors
	fromCell, toCell = toCell, fromCell

	var travelPath = g.TravelPath()
	var solvePath = g.SolvePath()
	var facing string = "north"

	// Get all distances from this cell
	d := fromCell.Distances()

	currentCell := toCell

	segment := maze.NewSegment(toCell, facing)
	travelPath.AddSegement(segment)
	solvePath.AddSegement(segment)

	for currentCell != d.Root() {
		// animation delay
		time.Sleep(delay)

		currentCell.SetVisited()

		smallest := math.MaxInt64
		var nextCell *maze.Cell
		for _, link := range currentCell.Links() {
			dist, _ := d.Get(link)
			if dist < smallest {
				smallest = dist
				nextCell = link
			}
		}

		facing = currentCell.GetFacingDirection(nextCell)

		segment := maze.NewSegment(nextCell, facing)
		travelPath.AddSegement(segment)
		solvePath.AddSegement(segment)

		g.SetPathFromTo(fromCell, currentCell, travelPath)
		currentCell = nextCell
	}

	// add toCell to path
	travelPath.ReverseCells()
	facing = currentCell.GetFacingDirection(toCell)

	segment = maze.NewSegment(toCell, facing)
	travelPath.AddSegement(segment)
	solvePath.AddSegement(segment)

	g.SetPathFromTo(fromCell, toCell, travelPath)

	// stats
	a.SetSolvePath(solvePath)
	a.SetTravelPath(travelPath)
	a.SetSolveSteps(solvePath.Length())

	return g, nil
}
