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
func randomUnvisitedDirection(d []*pb.Direction, v map[string]bool) string {
	// TODO(dan): Fix this to keep track of visited locations
	return d[utils.Random(0, len(d))].GetName()
}

func (a *RandomUnvisited) Solve(mazeID, clientID string, fromCell, toCell *pb.MazeLocation, delay time.Duration, directions []*pb.Direction) error {
	defer solvealgos.TimeTrack(a, time.Now())

	currentCell := fromCell
	solved := false

	// keep track of visited cells
	visited := make(map[string]bool)

	for !solved {
		// animation delay
		time.Sleep(delay)

		// mark cell as visited
		visited[currentCell.String()] = true

		if moveDir := randomUnvisitedDirection(directions, visited); moveDir != "" {
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
