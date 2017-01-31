// Package algos defines the interface for algorithms
package genalgos

import (
	"errors"
	"fmt"
	"mazes/grid"
)

type Algorithmer interface {
	Apply(*grid.Grid) (*grid.Grid, error)
	CheckGrid(*grid.Grid) error
}

type Common struct {
}

func (a *Common) Apply(*grid.Grid) (*grid.Grid, error) {
	return nil, errors.New("Apply() not implemented")
}

// CheckGrid checks that the generated grid is valid
func (a *Common) CheckGrid(g *grid.Grid) error {
	gWidth, gHeight := g.Dimensions()
	maxX, maxY := gWidth-1, gHeight-1

	for _, cell := range g.Cells() {
		// each cell must have at least one linked neighbor
		linksValid := false
		for _, n := range []*grid.Cell{cell.North, cell.East, cell.South, cell.West} {
			if n != nil {
				linksValid = true
				break // found a neighbor
			}
		}
		if !linksValid {
			return fmt.Errorf("cell %v does not have any open passages", cell)
		}

		// and between 2 and 4 total neighbors
		neighbors := cell.Neighbors()
		if len(neighbors) > 4 || len(neighbors) < 2 {
			return fmt.Errorf("cell %v has %v neighbors, this is not possible: %v", cell, len(neighbors), neighbors)

		}

		// walls
		if len(neighbors) == 3 {
			validLocations := []grid.Location{}

			// top and bottom rows
			for x := 1; x < maxX; x++ {
				for _, y := range []int{0, maxY} {
					validLocations = append(validLocations, grid.Location{x, y})
				}
			}

			// left and right columns
			for _, x := range []int{0, maxX} {
				for y := 1; y < maxY; y++ {
					validLocations = append(validLocations, grid.Location{x, y})
				}
			}

			cellLoc := cell.Location()

			if !grid.LocInLocList(cellLoc, validLocations) {
				return fmt.Errorf("cell %v is not a wall cell", cell)
			}
		}

		// corners
		if len(neighbors) == 2 {
			validLocations := []grid.Location{{0, 0}, {0, maxY}, {maxX, 0}, {maxX, maxY}}
			cellLoc := cell.Location()

			if !grid.LocInLocList(cellLoc, validLocations) {
				return fmt.Errorf("cell %v is not a corner cell", cell)
			}

		}
	}

	return nil
}
