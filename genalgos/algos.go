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
		return fmt.Errorf("tree node count != grid node count; tree=%v; grid=%v", t.NodeCount(), len(g.Cells()))
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
