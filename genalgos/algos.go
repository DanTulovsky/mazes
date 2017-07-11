// Package algos defines the interface for algorithms
package genalgos

import (
	"errors"
	"fmt"
	"log"
	"time"

	"mazes/maze"
	"mazes/tree"
	"mazes/utils"

	"github.com/tevino/abool"
)

type Algorithmer interface {
	Apply(m *maze.Maze, delay time.Duration, generating *abool.AtomicBool) error
	Cleanup(m *maze.Maze)
	CheckGrid(m *maze.Maze) error
}

type Common struct {
}

func (a *Common) Apply(m *maze.Maze, duration time.Duration, bool *abool.AtomicBool) (*maze.Maze, error) {
	return nil, errors.New("Apply() not implemented")
}

func Step(m *maze.Maze, t *tree.Tree, currentCell, parentCell *maze.Cell) bool {

	var nextCell *maze.Cell
	currentCell.SetVisited(maze.VisitedGenerator)

	if currentCell != parentCell {
		currentNode := tree.NewNode(currentCell.String())
		parentNode := t.Node(parentCell.String())

		t.AddNode(currentNode, parentNode)
	}

	// check for cycles
	for _, nextCell = range currentCell.Links() {
		if nextCell.Visited(maze.VisitedGenerator) {
			currentNode := t.Node(currentCell.String())
			nextNode := t.Node(nextCell.String())

			if nextNode == nil {
				// something is really wrong and should never happen
				maze.Fail(fmt.Errorf("unable to find %v in tree", nextCell))
			}

			if currentNode.Parent() != nextNode {
				maze.Fail(fmt.Errorf("found a cycle in the graph, %v is connected to %v, but %v is not the parent;\n%v", currentNode,
					nextNode, nextNode, t))
			}
		}

	}

	for _, nextCell = range currentCell.Links() {
		if !nextCell.Visited(maze.VisitedGenerator) {
			if Step(m, t, nextCell, currentCell) {
				return true
			}
		}

		currentCell.SetVisited(maze.VisitedGenerator)
	}

	return false
}

// CheckGrid checks that the generated grid is valid
func (a *Common) CheckGrid(m *maze.Maze) error {
	log.Print("Checking for cycles and converting to a spanning tree...")
	m.Reset()

	// convert grid to a spanning tree and check for cycles

	// We do a DFS traversal of the given graph. For every visited vertex ‘v’, if there is an adjacent ‘u’
	// such that u is already visited and u is not parent of v, then there is a cycle in graph. If we don’t
	// find such an adjacent for any vertex, we say that there is no cycle. The assumption of this approach
	// is that there are no parallel edges between any two vertices.

	start := m.RandomCell()
	rootNode := tree.NewNode(start.String())
	t, err := tree.NewTree(rootNode)
	if err != nil {
		return err
	}

	Step(m, t, start, start)

	// verify t has the same number of nodes as g
	if !m.Config().AllowWeaving {
		if t.NodeCount() != len(m.Cells()) {
			log.Printf("tree:\n%v\n", t)
			return fmt.Errorf("tree node count != grid node count; tree=%v; grid=%v", t.NodeCount(), len(m.Cells()))
		}
	}

	return nil
}

// Cleanup cleans up after generator is done
func (a *Common) Cleanup(m *maze.Maze) {
	m.SetGenCurrentLocation(nil)
}

func TimeTrack(m *maze.Maze, start time.Time) {
	m.SetCreateTime(time.Since(start))
}

// RandomUnvisitedCellFromList returns a random cell from n that has not been visited
func RandomUnvisitedCellFromList(neighbors []*maze.Cell) *maze.Cell {
	var allowed []*maze.Cell
	for _, n := range neighbors {
		if !n.Visited(maze.VisitedGenerator) {
			allowed = append(allowed, n)
		}
	}

	if len(allowed) == 0 {
		return nil
	}
	return allowed[utils.Random(0, len(allowed))]
}
