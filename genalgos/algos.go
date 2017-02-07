// Package algos defines the interface for algorithms
package genalgos

import (
	"errors"
	"fmt"
	"log"
	"mazes/grid"
	"mazes/tree"
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

func Step(g *grid.Grid, t *tree.Tree, currentCell, parentCell *grid.Cell) bool {

	var nextCell *grid.Cell
	currentCell.SetVisited()

	if currentCell != parentCell {
		currentNode := tree.NewNode(currentCell.String())
		parentNode := t.Node(parentCell.String())

		t.AddNode(currentNode, parentNode)
	}

	for _, nextCell = range currentCell.Links() {
		if !nextCell.Visited() {
			if Step(g, t, nextCell, currentCell) {
				return true
			}
		}

		currentCell.SetVisited()
	}

	return false
}

// CheckGrid checks that the generated grid is valid
func (a *Common) CheckGrid(g *grid.Grid) error {
	log.Print("Checking if grid converts to a spanning tree...")
	g.ResetVisited()

	// convert grid to a spanning tree
	start := g.RandomCell()
	rootNode := tree.NewNode(start.String())
	t, err := tree.NewTree(rootNode)
	if err != nil {
		return fmt.Errorf(err.Error())
	}

	Step(g, t, start, start)

	// verify t has the same number of nodes as g
	if t.NodeCount() != len(g.Cells()) {
		log.Printf("tree:\n%v\n", t)
		return fmt.Errorf("tree node count != grid node count; tree=%v; grid=%v", t.NodeCount(), len(g.Cells()))
	}

	log.Printf("\n%v\n", t)

	gWidth, gHeight := g.Dimensions()
	maxX, maxY := gWidth-1, gHeight-1

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
