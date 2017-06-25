// Package recursive_backtracker implements the recursive backtracker  maze solving algorithm
package recursive_backtracker

import (
	"fmt"
	"log"
	"time"

	"mazes/maze"
	pb "mazes/proto"
	"mazes/solvealgos"
)

var (
	steps int
)

type RecursiveBacktracker struct {
	solvealgos.Common
}

// Step steps into the next cell and returns true if it reach toCell.
func (a *RecursiveBacktracker) Step(mazeID, clientID string, currentCell, previousCell *pb.MazeLocation,
	directions []*pb.Direction, solved bool, delay time.Duration, m *maze.Maze) bool {
	// animation delay
	time.Sleep(delay)

	// set current location in local maze
	steps++
	if err := a.UpdateClientViewAndLocation(clientID, m, currentCell, previousCell, steps); err != nil {
		return false
	}
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
			previousCell = currentCell

			solved = reply.GetSolved()

			if a.Step(mazeID, clientID, reply.GetCurrentLocation(), previousCell, directions, solved, delay, m) {
				return true
			}

			time.Sleep(delay)
		}
	}

	reply, err := a.MoveBack(mazeID, clientID)
	if err != nil {
		log.Printf("error moving: %v", err)
		return false
	}

	previousCell = currentCell
	currentCell = reply.GetCurrentLocation()
	// set current location in local maze
	steps++
	if err := a.UpdateClientViewAndLocation(clientID, m, currentCell, previousCell, steps); err != nil {
		return false
	}
	return false
}

func (a *RecursiveBacktracker) Solve(mazeID, clientID string, fromCell, toCell *pb.MazeLocation,
	delay time.Duration, directions []*pb.Direction, m *maze.Maze) error {
	defer solvealgos.TimeTrack(a, time.Now())

	// DFS traversal of the grid
	if r := a.Step(mazeID, clientID, fromCell, nil, directions, false, delay, m); !r {
		return fmt.Errorf("failed to find path through maze from %v to %v", fromCell, toCell)
	}

	log.Printf("maze solved!")
	a.ShowStats()

	return nil
}
