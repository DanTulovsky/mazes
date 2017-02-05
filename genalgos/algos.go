// Package algos defines the interface for algorithms
package genalgos

import (
	"errors"
	"fmt"
	"mazes/grid"
	"mazes/utils"
	"time"
)

type Algorithmer interface {
	Apply(g *grid.Grid, delay time.Duration) (*grid.Grid, error)
	Cleanup(g *grid.Grid)
	CheckGrid(g *grid.Grid) error
}

type Common struct {
}

func (a *Common) Apply(*grid.Grid) (*grid.Grid, error) {
	return nil, errors.New("Apply() not implemented")
}

// CheckGrid checks that the generated grid is valid
func (a *Common) CheckGrid(g *grid.Grid) error {
	// gWidth, gHeight := g.Dimensions()
	// maxX, maxY := gWidth-1, gHeight-1

	for _, cell := range g.Cells() {
		// each cell must have at least one linked neighbor
		links := 0
		for _, n := range cell.Links() {
			if n != nil {
				links++
			}
		}
		if links < 0 || links > 4 {
			return fmt.Errorf("cell %v has invalid number of links: %v", cell, links)
		}

		// and between 1 and 4 total neighbors
		neighbors := cell.Neighbors()
		if len(neighbors) > 4 || len(neighbors) < 1 {
			return fmt.Errorf("cell %v has %v neighbors, this is not possible: %v", cell, len(neighbors), neighbors)

		}

		// These checks don't work for odd mazes
		// TODO(dan): Replace with spanning tree check
		// walls
		//if len(neighbors) == 3 {
		//	validLocations := []grid.Location{}
		//
		//	// top and bottom rows
		//	for x := 1; x < maxX; x++ {
		//		for _, y := range []int{0, maxY} {
		//			validLocations = append(validLocations, grid.Location{x, y})
		//		}
		//	}
		//
		//	// left and right columns
		//	for _, x := range []int{0, maxX} {
		//		for y := 1; y < maxY; y++ {
		//			validLocations = append(validLocations, grid.Location{x, y})
		//		}
		//	}
		//
		//	cellLoc := cell.Location()
		//
		//	if !grid.LocInLocList(cellLoc, validLocations) {
		//		return fmt.Errorf("cell %v is not a wall cell", cell)
		//	}
		//}
		//
		//// corners
		//if len(neighbors) == 2 {
		//	validLocations := []grid.Location{{0, 0}, {0, maxY}, {maxX, 0}, {maxX, maxY}}
		//	cellLoc := cell.Location()
		//
		//	if !grid.LocInLocList(cellLoc, validLocations) {
		//		return fmt.Errorf("cell %v is not a corner cell", cell)
		//	}
		//
		//}
	}

	return nil
}

// Cleanup cleans up after generator is done
func (a *Common) Cleanup(g *grid.Grid) {
	g.SetGenCurrentLocation(nil)
}

func TimeTrack(g *grid.Grid, start time.Time) {
	g.SetCreateTime(time.Since(start))
}

// RandomUnvisitedCellFromList returns a random cell from n that has not been visited
func RandomUnvisitedCellFromList(neighbors []*grid.Cell) *grid.Cell {
	var allowed []*grid.Cell
	for _, n := range neighbors {
		if !n.Visited() {
			allowed = append(allowed, n)
		}
	}

	if len(allowed) == 0 {
		return nil
	}
	return allowed[utils.Random(0, len(allowed))]
}
