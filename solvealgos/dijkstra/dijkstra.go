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

	var travelPath = maze.NewPath()
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
	travelPath.AddSegement(segment)
	solvePath.AddSegement(segment)

	for currentCell != d.Root() {
		// animation delay
		time.Sleep(delay)

		//currentCell.SetVisited()

		smallest := math.MaxInt64
		var nextCell *maze.Cell
		log.Printf("currentCell: %v", currentCell)
		// TODO: cells are missing links
		log.Printf("  links: %v", currentCell.Links())
		for _, link := range currentCell.Links() {
			log.Printf("  link: %v", link)
			dist, _ := d.Get(link)
			if dist < smallest {
				smallest = dist
				nextCell = link
			}
		}

		facing = currentCell.GetFacingDirection(nextCell)

		segment := maze.NewSegment(nextCell, facing, false)
		travelPath.AddSegement(segment)
		solvePath.AddSegement(segment)

		//m.SetClientPath(mazeFromCell, currentCell, travelPath)
		currentCell = nextCell
	}

	// add toCell to path
	travelPath.ReverseCells()
	facing = currentCell.GetFacingDirection(mazeToCell)

	segment = maze.NewSegment(mazeToCell, facing, false)
	travelPath.AddSegement(segment)
	solvePath.AddSegement(segment)

	//m.SetClientPath(mazeFromCell, mazeToCell, travelPath)

	// stats
	//a.SetSolvePath(solvePath)
	//a.SetTravelPath(travelPath)
	//a.SetSolveSteps(solvePath.Length())

	solved := false
	steps := 0
	currentServerCell := fromCell

	for _, seg := range solvePath.Segments() {
		// animation delay
		time.Sleep(delay)

		reply, err := a.Move(mazeID, clientID, seg.Cell().String())
		if err != nil {
			return err
		}
		//directions = reply.GetAvailableDirections()
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

	log.Printf("maze solved!")
	a.ShowStats()

	return nil
}
