// Package dijkstra implements dijkstra's algorithm to find the shortest path
// Note that this algorithm knows the entire layout of the maze and the distances between all cells
package dijkstra

import (
	pb "github.com/DanTulovsky/mazes/proto"
	"log"
	"math"
	"time"

	"github.com/DanTulovsky/mazes/maze"
	"github.com/DanTulovsky/mazes/solvealgos"
)

type Dijkstra struct {
	solvealgos.Common
}

func (a *Dijkstra) Solve(mazeID, clientID string, fromCell, toCell *pb.MazeLocation, delay time.Duration, _ []*pb.Direction, m *maze.Maze) error {
	defer solvealgos.TimeTrack(a, time.Now())

	// swap these for proper drawing colors
	fromCell, toCell = toCell, fromCell

	var solvePath = maze.NewPath()
	var facing = "north"

	mazeFromCell, err := m.CellFromLocation(fromCell)
	if err != nil {
		return err
	}
	mazeToCell, err := m.CellFromLocation(toCell)
	if err != nil {
		return err
	}

	// Get all distances from this cell
	d := mazeFromCell.Distances()

	currentCell := mazeToCell

	segment := maze.NewSegment(mazeToCell, facing, false)
	solvePath.AddSegement(segment)

	// Solve the maze locally, we know its entire layout
	for currentCell != d.Root() {
		smallest := math.MaxInt64
		var nextCell *maze.Cell
		for _, link := range currentCell.Links() {
			dist, _ := d.Get(link)
			if dist < smallest {
				smallest = dist
				nextCell = link
			}
		}

		facing = currentCell.GetFacingDirection(nextCell)
		segment := maze.NewSegment(nextCell, facing, false)
		solvePath.AddSegement(segment)

		currentCell = nextCell
	}

	// add toCell to path
	facing = currentCell.GetFacingDirection(mazeToCell)

	segment = maze.NewSegment(mazeToCell, facing, false)
	solvePath.AddSegement(segment)

	solved := false
	steps := 0
	currentServerCell := fromCell

	s := solvePath.Segments()
	for i := 1; i < len(s); i++ {
		// animation delay
		time.Sleep(delay)

		direction, err := s[i-1].Cell().DirectionTo(s[i].Cell(), clientID)
		if err != nil {
			return err
		}

		reply, err := a.Move(mazeID, clientID, direction.Name)
		if err != nil {
			return err
		}

		previousServerCell := currentServerCell
		currentServerCell = reply.GetCurrentLocation()

		// set current location in local maze
		steps++
		if err := a.UpdateClientViewAndLocation(clientID, m, currentServerCell, previousServerCell, steps); err != nil {
			return err
		}

		solved = reply.Solved
		if solved {
			break
		}
	}

	log.Printf("maze solved in %v steps", steps)
	a.ShowStats()

	return nil
}
