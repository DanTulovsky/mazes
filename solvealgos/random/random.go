// Package random implements the random walk maze solving algorithm

// Walk around the maze until you find a solution.  Dumb as it gets.
package random

import (
	"fmt"
	"log"
	"time"

	pb "mazes/proto"
	"mazes/solvealgos"
	"mazes/utils"
)

type Random struct {
	solvealgos.Common
}

// randomDirection returns a random direction from the list of available ones
func randomDirection(d []*pb.Direction) string {
	return d[utils.Random(0, len(d))].GetName()
}

func (a *Random) Solve(mazeID, clientID string, fromCell, toCell *pb.MazeLocation, delay time.Duration, directions []*pb.Direction) error {
	defer solvealgos.TimeTrack(a, time.Now())

	currentCell := fromCell
	solved := false

	for !solved {
		// animation delay
		time.Sleep(delay)

		if nextCell := randomDirection(directions); nextCell != "" {
			reply, err := a.Move(mazeID, clientID, nextCell)
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
