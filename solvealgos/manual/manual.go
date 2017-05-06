package manual

import (
	"errors"
	"fmt"
	"log"
	"mazes/maze"
	"mazes/solvealgos"
	"strings"
	"time"
)

type Manual struct {
	solvealgos.Common
}

func getNextCell(currentCell *maze.Cell, key string) (*maze.Cell, error) {
	switch key {
	case "Up":
		if currentCell.Linked(currentCell.North()) {
			return currentCell.North(), nil
		}
	case "Down":
		if currentCell.Linked(currentCell.South()) {
			return currentCell.South(), nil
		}
	case "Left":
		if currentCell.Linked(currentCell.West()) {
			return currentCell.West(), nil
		}
	case "Right":
		if currentCell.Linked(currentCell.East()) {
			return currentCell.East(), nil
		}
	}
	return nil, fmt.Errorf("unable to move %s, no passage in that direction", key)

}

func (a *Manual) Solve(m *maze.Maze, fromCell, toCell *maze.Cell, delay time.Duration, keyInput <-chan string) (*maze.Maze, error) {
	defer solvealgos.TimeTrack(a, time.Now())

	log.Print("Solver is human...")
	var travelPath = m.TravelPath()
	var solvePath = m.SolvePath()

	currentCell := fromCell
	facing := "north"

	// Visit cells
	for currentCell != toCell {
		// TODO(dant): Separate travelPath and solvePath
		currentCell.SetVisited()
		segment := maze.NewSegment(currentCell, facing)
		travelPath.AddSegement(segment)
		solvePath.AddSegement(segment)

		m.SetPathFromTo(fromCell, currentCell, travelPath)

		var nextCell *maze.Cell
		var err error = errors.New("no new cell yet")

		for err != nil {
			// get nextCell from user input based on key press
			key := <-keyInput
			switch strings.ToLower(key) {
			case "q":
				return m, errors.New("received cancel request, exiting...")
			default:
				nextCell, err = getNextCell(currentCell, key)

			}

			if err != nil {
				log.Printf("cannot move %s: %v", key, err)
				continue
			}
		}
		facing = currentCell.GetFacingDirection(nextCell)
		currentCell = nextCell
	}

	// add the last cell
	facing = currentCell.GetFacingDirection(toCell)
	segment := maze.NewSegment(toCell, facing)
	travelPath.AddSegement(segment)
	solvePath.AddSegement(segment)
	m.SetPathFromTo(fromCell, toCell, travelPath)

	// Set final paths
	a.SetSolvePath(solvePath)
	a.SetTravelPath(travelPath)
	a.SetSolveSteps(travelPath.Length())

	return m, nil
}
