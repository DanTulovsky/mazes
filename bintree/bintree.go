package bintree

import "mazes/grid"

// Apply applies the binary tree algorithm to the grid.
func Apply(g *grid.Grid) *grid.Grid {

	for _, cell := range g.Cells() {
		neighbors := []*grid.Cell{}
		if cell.North != nil {
			neighbors = append(neighbors, cell.North)
		}
		if cell.East != nil {
			neighbors = append(neighbors, cell.East)
		}

		if len(neighbors) == 0 {
			continue
		}
		index := grid.Random(0, len(neighbors))
		neighbor := neighbors[index]
		if neighbor != nil {
			cell.Link(neighbor)
		}
	}
	return g
}
