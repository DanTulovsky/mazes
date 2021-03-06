// Package random_unvisited implements the random unvisited walk maze solving algorithm

// Walk around the maze until you find a solution. Prefer unvisited first.
package random_unvisited

import (
	"fmt"
	"log"
	"time"

	"github.com/DanTulovsky/mazes/maze"
	pb "github.com/DanTulovsky/mazes/proto"
	"github.com/DanTulovsky/mazes/solvealgos"
	"github.com/DanTulovsky/mazes/utils"
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

func (a *RandomUnvisited) Solve(mazeID, clientID string, fromCell, toCell *pb.MazeLocation,
	delay time.Duration, directions []*pb.Direction, m *maze.Maze) error {
	defer solvealgos.TimeTrack(a, time.Now())

	currentCell := fromCell
	solved := false
	steps := 0

	for !solved {
		// animation delay
		time.Sleep(delay)

		if moveDir := randomUnvisitedDirection(directions); moveDir != "" {
			reply, err := a.Move(mazeID, clientID, moveDir)
			if err != nil {
				return err
			}
			directions = reply.GetAvailableDirections()
			previousCell := currentCell
			currentCell = reply.GetCurrentLocation()

			// set current location in local maze
			steps++
			if err := a.UpdateClientViewAndLocation(clientID, m, currentCell, previousCell, steps); err != nil {
				return err
			}
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
