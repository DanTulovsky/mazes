// Package hint_and_kill implements the hunt-and-Kill algorithm for maze generation

package hint_and_kill

import (
	"math/rand"
	"mazes/genalgos"
	"mazes/maze"
	"time"
)

type HuntAndKill struct {
	genalgos.Common
}

// Hunt scans the grid from left to right and returns the first unvisited cell with at least one visited neighbor
// Returns nil if there are no more
func HuntAndLink(g *maze.Grid) *maze.Cell {
	for _, cell := range g.Cells() {
		if cell.Visited() {
			continue
		}
		// shuffle the neighbors so we get a random one for linking
		for _, n := range Shuffle(cell.Neighbors()) {
			if n.Visited() {
				cell.Link(n) // link to random neighbor
				return cell
			}
		}
	}
	return nil
}

func Shuffle(cells []*maze.Cell) []*maze.Cell {
	for i := range cells {
		j := rand.Intn(i + 1)
		cells[i], cells[j] = cells[j], cells[i]
	}
	return cells
}

// Apply applies the binary tree algorithm to generate the maze.
func (a *HuntAndKill) Apply(g *maze.Grid, delay time.Duration) (*maze.Grid, error) {

	defer genalgos.TimeTrack(g, time.Now())

	currentCell := g.RandomCell()

	for currentCell != nil {
		time.Sleep(delay) // animation delay
		g.SetGenCurrentLocation(currentCell)

		currentCell.SetVisited()
		neighbors := currentCell.Neighbors()

		randomNeighbor := genalgos.RandomUnvisitedCellFromList(neighbors)
		if randomNeighbor == nil {
			// no more unvisited neighbors
			currentCell = HuntAndLink(g)
			continue
		}

		currentCell.Link(randomNeighbor)
		currentCell = randomNeighbor
	}

	a.Cleanup(g)
	return g, nil
}
