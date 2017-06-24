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

type RecursiveBacktracker struct {
	solvealgos.Common
}

// Step steps into the next cell and returns true if it reach toCell.
func (a *RecursiveBacktracker) Step(mazeID, clientID string, currentCell *pb.MazeLocation,
	directions []*pb.Direction, solved bool, delay time.Duration, m *maze.Maze) bool {
	// animation delay
	time.Sleep(delay)

	if solved {
		return true
	}

	client, err := m.Client(clientID)
	if err != nil {
		return false
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

			if cell, err := a.CellForLocation(m, currentCell); err != nil {
				return false
			} else {
				client.SetCurrentLocation(cell)
			}

			solved = reply.GetSolved()

			if a.Step(mazeID, clientID, currentCell, directions, solved, delay, m) {
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

	if cell, err := a.CellForLocation(m, reply.GetCurrentLocation()); err != nil {
		return false
	} else {
		client.SetCurrentLocation(cell)
	}

	return false
}

func (a *RecursiveBacktracker) Solve(mazeID, clientID string, fromCell, toCell *pb.MazeLocation,
	delay time.Duration, directions []*pb.Direction, m *maze.Maze) error {
	defer solvealgos.TimeTrack(a, time.Now())

	// DFS traversal of the grid
	if r := a.Step(mazeID, clientID, fromCell, directions, false, delay, m); !r {
		return fmt.Errorf("failed to find path through maze from %v to %v", fromCell, toCell)
	}

	log.Printf("maze solved!")
	a.ShowStats()

	return nil
}
