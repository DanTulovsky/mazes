// Package recursive_backtracker implements the recursive backtracker  maze solving algorithm
package recursive_backtracker

import (
	"fmt"
	"log"
	"time"

	pb "mazes/proto"
	"mazes/solvealgos"
)

type RecursiveBacktracker struct {
	solvealgos.Common
}

// Step steps into the next cell and returns true if it reach toCell.
func (a *RecursiveBacktracker) Step(mazeID, clientID string, currentCell *pb.MazeLocation, directions []*pb.Direction, solved bool, delay time.Duration) bool {
	// animation delay
	time.Sleep(delay)

	if solved {
		return true
	}

	for _, nextDir := range directions {
		if !nextDir.GetVisited() {
			reply, err := a.Move(mazeID, clientID, nextDir.GetName())
			if err != nil {
				log.Printf("error moving: %v", err)
				return false
			}
			directions = reply.GetAvailableDirections()
			currentCell = reply.GetCurrentLocation()
			solved = reply.GetSolved()

			if a.Step(mazeID, clientID, currentCell, directions, solved, delay) {
				return true
			}

			time.Sleep(delay)
		}
	}

	a.MoveBack(mazeID, clientID)
	return false
}

func (a *RecursiveBacktracker) Solve(mazeID, clientID string, fromCell, toCell *pb.MazeLocation, delay time.Duration, directions []*pb.Direction) error {
	defer solvealgos.TimeTrack(a, time.Now())

	// DFS traversal of the grid
	if r := a.Step(mazeID, clientID, fromCell, directions, false, delay); !r {
		return fmt.Errorf("failed to find path through maze from %v to %v", fromCell, toCell)
	}

	log.Printf("maze solved!")
	a.ShowStats()

	return nil
}
