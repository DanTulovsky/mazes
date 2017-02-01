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

// getDirections returns the possible directions to move in the proper order based on which way you are "facing"
func getDirections(c *grid.Cell, facing string) []*grid.Cell {

	switch facing {
	case "north":
		return []*grid.Cell{c.East, c.North, c.West, c.South}
	case "east":
		return []*grid.Cell{c.South, c.East, c.North, c.West}
	case "south":
		return []*grid.Cell{c.West, c.South, c.East, c.North}
	case "west":
		return []*grid.Cell{c.North, c.West, c.South, c.East}
	}
	return nil
}

func pickNextCell(currentCell *grid.Cell, facing string) *grid.Cell {
	// always go in this order: "right", "forward", "left", "back"

	dirs := getDirections(currentCell, facing)

	for _, l := range dirs {
		if currentCell.Linked(l) {
			return l
		}
	}
	// backtrack if we can't go anywhere else
	return nil
}

func (a *WallFollower) Solve(g *grid.Grid, fromCell, toCell *grid.Cell) (*grid.Grid, error) {
	defer solvealgos.TimeTrack(a, time.Now())

	var path = grid.NewStack()

	currentCell := fromCell
	facing := "north"

	log.Printf("%v -> %v", fromCell, toCell)

	for currentCell != toCell {
		path.Push(currentCell)
		log.Printf("facing: %v", facing)

		if nextCell := pickNextCell(currentCell, facing); nextCell != nil {
			if currentCell.North == nextCell {
				facing = "north"
			}
			if currentCell.East == nextCell {
				facing = "east"
			}
			if currentCell.West == nextCell {
				facing = "west"
			}
			if currentCell.South == nextCell {
				facing = "south"
			}

			log.Printf("path: %v", path)
			currentCell = nextCell
		} else {
			// this can never happen unless the maze is broken
			return nil, fmt.Errorf("%v isn't linked to any other cell, failing", currentCell)

		}
	}

	path.Push(toCell)
	log.Printf("final path: %v", path)
	g.SetPathFromTo(fromCell, toCell, path.List())
	// stats
	a.SetSolvePath(path.List())
	a.SetSolveSteps(len(path.List())) // always the same as the actual path

	return g, nil
}
