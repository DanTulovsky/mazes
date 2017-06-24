// Package wall_follower implements the wall follower maze solving algorithm

//  Start following passages, and whenever you reach a junction always turn right (or left).
// Equivalent to a human solving a Maze by putting their hand on the right (or left) wall and
// leaving it there as they walk through.
package wall_follower

import (
	"fmt"
	"log"
	"time"

	"mazes/maze"
	pb "mazes/proto"
	"mazes/solvealgos"
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
	return []string{}
}

func pickNextDir(directions []*pb.Direction, facing string) string {
	// always go in this order: "right", "forward", "left", "back"

	dirs := getDirections(facing)
	if len(dirs) == 0 {
		return ""
	}

	for _, l := range dirs {
		for _, d := range directions {
			if d.GetName() == l {
				return l
			}
		}
	}
	return ""
}

func (a *WallFollower) Solve(mazeID, clientID string, fromCell, toCell *pb.MazeLocation,
	delay time.Duration, directions []*pb.Direction, m *maze.Maze) error {
	defer solvealgos.TimeTrack(a, time.Now())

	if len(directions) < 1 {
		return fmt.Errorf("no available directions to move: %v", directions)
	}

	currentCell := fromCell
	client, err := m.Client(clientID)
	if err != nil {
		return err
	}

	facing := directions[0].GetName()
	solved := false

	// keep track of how many times each cell has been visited
	visited := make(map[string]int)

	for !solved {
		// animation delay
		time.Sleep(delay)

		if _, ok := visited[currentCell.String()]; !ok {
			visited[currentCell.String()] = 0
		}
		visited[currentCell.String()]++

		if visited[currentCell.String()] > 4 {
			// we are stuck in a loop, fail
			return fmt.Errorf("cell %v visited %v times, stuck in a loop", currentCell.String(), visited[currentCell.String()])
		}

		if nextCell := pickNextDir(directions, facing); nextCell != "" {
			facing = nextCell
			reply, err := a.Move(mazeID, clientID, nextCell)
			if err != nil {
				return err
			}
			directions = reply.GetAvailableDirections()
			currentCell = reply.GetCurrentLocation()

			if cell, err := a.CellForLocation(m, currentCell); err != nil {
				return err
			} else {
				client.SetCurrentLocation(cell)
			}

			solved = reply.Solved
		} else {
			// this can never happen unless the maze is broken
			return fmt.Errorf("%v isn't linked to any other cell, failing", currentCell)

		}
	}

	log.Printf("maze solved!")
	a.ShowStats()

	return nil
}
