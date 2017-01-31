// Package hint_and_kill implements the hunt-and-Kill algorithm for maze generation

package hint_and_kill

import (
	"math/rand"
	"mazes/genalgos"
	"mazes/grid"
	"mazes/utils"
	"time"
)

type HuntAndKill struct {
	genalgos.Common
}

// RandomUnvisitedCellFromList returns a random cell from n that is not in visited
func RandomUnvisitedCellFromList(neighbors []*grid.Cell, visited []*grid.Cell) *grid.Cell {
	var allowed []*grid.Cell
	for _, n := range neighbors {
		if !grid.CellInCellList(n, visited) {
			allowed = append(allowed, n)
		}
	}

	if len(allowed) == 0 {
		return nil
	}
	return allowed[utils.Random(0, len(allowed))]
}

// Hunt scans the grid from left to right and returns the first unvisited cell with at least one visited neighbor
// Returns nil if there are no more
func HuntAndLink(g *grid.Grid) *grid.Cell {
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

func Shuffle(cells []*grid.Cell) []*grid.Cell {
	for i := range cells {
		j := rand.Intn(i + 1)
		cells[i], cells[j] = cells[j], cells[i]
	}
	return cells
}

// Apply applies the binary tree algorithm to generate the maze.
func (a *HuntAndKill) Apply(g *grid.Grid) (*grid.Grid, error) {

	defer utils.TimeTrack(time.Now(), "hunt-and-kill apply")

	var visitedCells []*grid.Cell
	currentCell := g.RandomCell()

	for currentCell != nil {
		visitedCells = append(visitedCells, currentCell)
		currentCell.SetVisited()
		neighbors := currentCell.Neighbors()

		randomNeighbor := RandomUnvisitedCellFromList(neighbors, visitedCells)
		if randomNeighbor == nil {
			// no more unvisited neighbors
			currentCell = HuntAndLink(g)
			continue
		}

		currentCell.Link(randomNeighbor)
		currentCell = randomNeighbor
	}
	return g, nil

}
