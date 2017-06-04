// Package random_unvisited implements the random unvisited walk maze solving algorithm

// Walk around the maze until you find a solution. Prefer unvisited first.
package random_unvisited

import (
	"fmt"
	"log"
	"time"

	pb "mazes/proto"
	"mazes/solvealgos"
	"mazes/utils"
)

type RandomUnvisited struct {
	solvealgos.Common
}

// randomDirection returns a random direction from the list of available ones
func randomUnvisitedDirection(directions []*pb.Direction) string {
	available := []*pb.Direction{}
	for _, dir := range directions {
		if !dir.GetVisited() {
			available = append(available, dir)
		}
	}

	if len(available) > 0 {
		return available[utils.Random(0, len(available))].GetName()
	}

	return directions[utils.Random(0, len(directions))].GetName()
}

func (a *RandomUnvisited) Solve(mazeID, clientID string, fromCell, toCell *pb.MazeLocation, delay time.Duration, directions []*pb.Direction) error {
	defer solvealgos.TimeTrack(a, time.Now())

	currentCell := fromCell
	solved := false

	for !solved {
		// animation delay
		time.Sleep(delay)

		if moveDir := randomUnvisitedDirection(directions); moveDir != "" {
			reply, err := a.Move(mazeID, clientID, moveDir)
			if err != nil {
				return err
			}
			directions = reply.GetAvailableDirections()
			currentCell = reply.GetCurrentLocation()
			solved = reply.Solved
		} else {
			// nowhere to go?
			return fmt.Errorf("%v isn't linked to any other cell, failing", currentCell)

		}
	}

	log.Printf("maze solved!")
	a.ShowStats()

	return nil
}
