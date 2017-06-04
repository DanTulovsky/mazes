// Package recursive_backtracker implements the recursive backtracker  maze solving algorithm
package recursive_backtracker

import (
	"fmt"
	"log"
	"time"

	pb "mazes/proto"
	"mazes/solvealgos"
)

// map of MazeLocation -> list of directions we've gone from that location
var visited map[string][]string

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
		if !nextDir.Visited {
			reply, err := a.Move(mazeID, clientID, nextDir.GetName())
			if err != nil {
				log.Printf("error moving: %v", err)
				return false
			}
			directions = reply.GetAvailableDirections()
			currentCell = reply.GetCurrentLocation()
			solved = reply.Solved

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

	solved := false
	visited = make(map[string][]string)

	// DFS traversal of the grid
	if r := a.Step(mazeID, clientID, fromCell, directions, solved, delay); !r {
		return fmt.Errorf("failed to find path through maze from %v to %v", fromCell, toCell)
	}

	log.Printf("maze solved!")
	a.ShowStats()

	return nil
}
