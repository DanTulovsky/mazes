// Package random implements the random walk maze solving algorithm

// Walk around the maze until you find a solution.  Dumb as it gets.
package random

import (
	"errors"
	"log"
	"mazes/maze"
	"mazes/solvealgos"
	"strings"
	"time"
)

type Random struct {
	solvealgos.Common
}

func (a *Random) Solve(m *maze.Maze, fromCell, toCell *maze.Cell, delay time.Duration, keyInput <-chan string) (*maze.Maze, error) {
	defer solvealgos.TimeTrack(a, time.Now())

	var travelPath = m.TravelPath()
	var solvePath = m.SolvePath()
	currentCell := fromCell
	facing := "north" // arbitrary

	for currentCell != toCell {
		// animation delay
		time.Sleep(delay)

		currentCell.SetVisited()

		segment := maze.NewSegment(currentCell, facing)
		travelPath.AddSegement(segment)
		solvePath.AddSegement(segment)
		m.SetPathFromTo(fromCell, currentCell, travelPath)

		nextCell := currentCell.RandomLink()
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

	// add the last cell
	facing = currentCell.GetFacingDirection(toCell)
	segment := maze.NewSegment(toCell, facing)
	travelPath.AddSegement(segment)
	solvePath.AddSegement(segment)
	m.SetPathFromTo(fromCell, toCell, travelPath)

	// stats
	a.SetSolvePath(solvePath)
	a.SetTravelPath(travelPath)
	a.SetSolveSteps(travelPath.Length())

	return m, nil
}
