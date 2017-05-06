// Package random_unvisited implements the random unvisited walk maze solving algorithm

// Walk around the maze until you find a solution. Prefer unvisited first.
package random_unvisited

import (
	"errors"
	"log"
	"mazes/maze"
	"mazes/solvealgos"
	"strings"
	"time"
)

type RandomUnvisited struct {
	solvealgos.Common
}

func (a *RandomUnvisited) Solve(m *maze.Maze, fromCell, toCell *maze.Cell, delay time.Duration, keyInput <-chan string) (*maze.Maze, error) {
	defer solvealgos.TimeTrack(a, time.Now())

	var travelPath = m.TravelPath()
	var solvePath = m.SolvePath()
	currentCell := fromCell
	facing := "north"

	for currentCell != toCell {
		// animation delay
		time.Sleep(delay)

		currentCell.SetVisited()

		segment := maze.NewSegment(currentCell, facing)
		travelPath.AddSegement(segment)
		solvePath.AddSegement(segment)
		m.SetPathFromTo(fromCell, currentCell, travelPath)

		// prefer unvisited first
		nextCell := currentCell.RandomUnvisitedLink()

		if nextCell == nil {
			nextCell = currentCell.RandomLink()
		}

		facing = currentCell.GetFacingDirection(nextCell)
		currentCell = nextCell

		select {
		case key := <-keyInput:
			switch strings.ToLower(key) {
			case "q":
				log.Print("Exiting...")
				return m, errors.New("received cancel request, exiting...")
			}
		default:
			// fmt.Println("no message received")
		}

	}

	// last cell
	facing = currentCell.GetFacingDirection(toCell)
	segment := maze.NewSegment(currentCell, facing)
	travelPath.AddSegement(segment)
	solvePath.AddSegement(segment)
	m.SetPathFromTo(fromCell, toCell, solvePath)

	// stats
	a.SetSolvePath(solvePath)
	a.SetTravelPath(travelPath)
	a.SetSolveSteps(travelPath.Length())

	return m, nil
}
