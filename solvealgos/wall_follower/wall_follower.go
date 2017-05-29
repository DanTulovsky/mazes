// Package wall_follower implements the wall follower maze solving algorithm

//  Start following passages, and whenever you reach a junction always turn right (or left).
// Equivalent to a human solving a Maze by putting their hand on the right (or left) wall and
// leaving it there as they walk through.
package wall_follower

import (
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"mazes/maze"
	pb "mazes/proto"
	"mazes/solvealgos"
)

type WallFollower struct {
	solvealgos.Common
}

// getDirections returns the possible directions to move in the proper order based on which way you are "facing"
func getDirections(c *maze.Cell, facing string) []*maze.Cell {

	switch facing {
	case "north":
		return []*maze.Cell{c.East(), c.North(), c.West(), c.South()}
	case "east":
		return []*maze.Cell{c.South(), c.East(), c.North(), c.West()}
	case "south":
		return []*maze.Cell{c.West(), c.South(), c.East(), c.North()}
	case "west":
		return []*maze.Cell{c.North(), c.West(), c.South(), c.East()}
	}
	return nil
}

func pickNextCell(currentCell *maze.Cell, facing string) *maze.Cell {
	// always go in this order: "right", "forward", "left", "back"

	dirs := getDirections(currentCell, facing)
	if dirs == nil {
		return nil
	}

	for _, l := range dirs {
		if currentCell.Linked(l) {
			return l
		}
	}
	return nil
}

func (a *WallFollower) Solve(stream pb.Mazer_SolveMazeClient, fromCell, toCell string, delay time.Duration) error {
	defer solvealgos.TimeTrack(a, time.Now())

	var travelPath = m.TravelPath()
	var solvePath = m.SolvePath()

	currentCell := fromCell
	facing := "north"

	for currentCell != toCell {
		// animation delay
		time.Sleep(delay)
		// this stuff happens on the server now
		currentCell.SetVisited()

		segment := maze.NewSegment(currentCell, facing)
		travelPath.AddSegement(segment)
		solvePath.AddSegement(segment)
		m.SetPathFromTo(fromCell, currentCell, travelPath)

		if currentCell.VisitedTimes() > 4 {
			// we are stuck in a loop, fail
			return nil, fmt.Errorf("cell %v visited %v times, stuck in a loop", currentCell, currentCell.VisitedTimes())
		}

		if nextCell := pickNextCell(currentCell, facing); nextCell != nil {
			if currentCell.North() == nextCell {
				facing = "north"
			}
			if currentCell.East() == nextCell {
				facing = "east"
			}
			if currentCell.West() == nextCell {
				facing = "west"
			}
			if currentCell.South() == nextCell {
				facing = "south"
			}

			currentCell = nextCell
		} else {
			// this can never happen unless the maze is broken
			return nil, fmt.Errorf("%v isn't linked to any other cell, failing", currentCell)

		}

		select {
		case key := <-keyInput:
			switch strings.ToLower(key) {
			case "q":
				log.Print("Exiting...")
				return m, errors.New("received cancel request, exiting...")
			}
		default:
			// fmt.Println("no message received")
		}
	}

	// last cell
	segment := maze.NewSegment(toCell, facing)
	travelPath.AddSegement(segment)
	solvePath.AddSegement(segment)
	m.SetPathFromTo(fromCell, toCell, solvePath)

	// stats
	a.SetSolvePath(solvePath)
	a.SetTravelPath(travelPath)
	a.SetSolveSteps(travelPath.Length()) // always the same as the actual path

	return m, nil
}
