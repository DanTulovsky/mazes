// Package solvealgos defines the interface for solver algorithms
package solvealgos

import (
	"errors"
	"mazes/grid"
)

type Algorithmer interface {
	Solve(*grid.Grid, *grid.Cell, *grid.Cell) (*grid.Grid, error)
}

type Common struct {
}

// Solve should write the path of the solution to the grid
func (a *Common) Solve(g *grid.Grid, fromCell, toCell *grid.Cell) (*grid.Grid, error) {
	return nil, errors.New("Solve() not implemented")
}
