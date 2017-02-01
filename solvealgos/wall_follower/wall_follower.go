// Package wall_follower implements the wall follower maze solving algorithm

//  Start following passages, and whenever you reach a junction always turn right (or left).
// Equivalent to a human solving a Maze by putting their hand on the right (or left) wall and
// leaving it there as they walk through.
package wall_follower

import (
	"fmt"
	"log"
	"mazes/grid"
	"mazes/solvealgos"
	"time"
)

type WallFollower struct {
	solvealgos.Common
}

func pickNextCell(currentCell, previousCell *grid.Cell) *grid.Cell {
	// always go in this order:  East -> North -> West -> South
	for _, l := range []*grid.Cell{currentCell.East, currentCell.North, currentCell.West, currentCell.South} {
		if currentCell.Linked(l) && l != previousCell {
			return l
		}
	}
	// only return previousCell if there were no other options
	return previousCell
}

func (a *WallFollower) Solve(g *grid.Grid, fromCell, toCell *grid.Cell) (*grid.Grid, error) {
	defer solvealgos.TimeTrack(a, time.Now())

	var path = grid.NewStack()

	currentCell := fromCell
	var previousCell *grid.Cell

	log.Printf("%v -> %v", fromCell, toCell)
	for currentCell != toCell {
		log.Printf("currentCell: %v", currentCell)
		path.Push(currentCell)
		if nextCell := pickNextCell(currentCell, previousCell); nextCell != nil {
			log.Printf("nextCell: %v", nextCell)
			previousCell = currentCell
			currentCell = nextCell
		} else {
			// this can never happen unless the maze is broken
			return nil, fmt.Errorf("%v isn't linked to any other cell, failing", currentCell)

		}
	}

	path.Push(toCell)
	g.SetPathFromTo(fromCell, toCell, path.List())
	// stats
	a.SetSolvePath(path.List())

	return g, nil
}
