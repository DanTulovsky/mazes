// Package wall_follower implements the wall follower maze solving algorithm

//  Start following passages, and whenever you reach a junction always turn right (or left).
// Equivalent to a human solving a Maze by putting their hand on the right (or left) wall and
// leaving it there as they walk through.
package wall_follower

import (
	"fmt"
	"log"
	"time"

	pb "mazes/proto"
	"mazes/solvealgos"
	"mazes/utils"
)

type WallFollower struct {
	solvealgos.Common
}

// getDirections returns the possible directions to move in the proper order based on which way you are "facing"
func getDirections(facing string) []string {

	switch facing {
	case "north":
		return []string{"east", "north", "west", "south"}
	case "east":
		return []string{"south", "east", "north", "west"}
	case "south":
		return []string{"west", "south", "east", "north"}
	case "west":
		return []string{"north", "west", "south", "east"}
	}
	return nil
}

func pickNextCell(directions []string, facing string) string {
	// always go in this order: "right", "forward", "left", "back"

	dirs := getDirections(facing)
	if dirs == nil {
		return nil
	}

	for _, l := range dirs {
		if utils.StrInList(directions, l) {
			return l
		}
	}
	return ""
}

func (a *WallFollower) Solve(stream pb.Mazer_SolveMazeClient, mazeID, clientID string, fromCell, toCell *pb.MazeLocation, delay time.Duration, directions []string) error {
	defer solvealgos.TimeTrack(a, time.Now())

	log.Printf("fromCell: %v; toCell: %v; directions: %v", fromCell, toCell, directions)
	if len(directions) < 1 {
		return fmt.Errorf("no available directions to move: %v", directions)
	}

	currentCell := fromCell
	facing := directions[0]

	// keep track of how many times each cell has been visited
	visited := make(map[string]int)

	for currentCell != toCell {
		// animation delay
		time.Sleep(delay)

		if _, ok := visited[currentCell.String()]; !ok {
			visited[currentCell.String()] = 0
		}
		visited[currentCell.String()]++

		if visited[currentCell.String()] > 4 {
			// we are stuck in a loop, fail
			return fmt.Errorf("cell %v visited %v times, stuck in a loop", currentCell, visited[currentCell.String()])
		}

		if nextCell := pickNextCell(currentCell, facing); nextCell != "" {
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
			return fmt.Errorf("%v isn't linked to any other cell, failing", currentCell)

		}

		//select {
		//case key := <-keyInput:
		//	switch strings.ToLower(key) {
		//	case "q":
		//		log.Print("Exiting...")
		//		return errors.New("received cancel request, exiting...")
		//	}
		//default:
		//	// fmt.Println("no message received")
		//}
	}

	// stats
	//a.SetSolvePath(solvePath)
	//a.SetTravelPath(travelPath)
	//a.SetSolveSteps(travelPath.Length()) // always the same as the actual path

	return nil
}
