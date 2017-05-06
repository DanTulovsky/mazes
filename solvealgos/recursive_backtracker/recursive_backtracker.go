// Package recursive_backtracker implements the recursive backtracker  maze solving algorithm

// TODO(dant): Fix me!
package recursive_backtracker

import (
	"fmt"
	"log"
	"mazes/maze"
	"mazes/solvealgos"
	"strings"
	"time"
)

var travelPath *maze.Path
var facing string = "north"
var startCell *maze.Cell

type RecursiveBacktracker struct {
	solvealgos.Common
}

// Step steps into the next cell and returns true if it reach toCell.
func Step(m *maze.Maze, currentCell, toCell *maze.Cell, solvePath *maze.Path, delay time.Duration, keyInput <-chan string) bool {
	// animation delay
	// log.Printf("currentCell: %v", currentCell)
	time.Sleep(delay)

	var nextCell *maze.Cell
	currentCell.SetVisited()

	segment := maze.NewSegment(currentCell, facing)
	solvePath.AddSegement(segment)
	travelPath.AddSegement(segment)
	m.SetPathFromTo(startCell, currentCell, travelPath)

	if currentCell == toCell {
		return true
	}

	for _, nextCell = range currentCell.Links() {
		if !nextCell.Visited() {
			facing = currentCell.GetFacingDirection(nextCell)
			segment.UpdateFacingDirection(facing)
			if Step(m, nextCell, toCell, solvePath, delay, keyInput) {
				return true
			}
		}

		facing = nextCell.GetFacingDirection(currentCell)

		// don't add the same segment if it's already the last one
		if travelPath.LastSegment().Cell() == currentCell {
			continue
		}

		segmentReturn := maze.NewSegment(currentCell, facing)
		travelPath.AddSegement(segmentReturn)
		currentCell.SetVisited()
		m.SetPathFromTo(startCell, currentCell, travelPath)

	}
	solvePath.DelSegement()
	time.Sleep(delay)

	select {
	case key := <-keyInput:
		switch strings.ToLower(key) {
		case "q":
			log.Print("Exiting...")
			return true
		}
	default:
		// fmt.Println("no message received")
	}

	return false
}

func (a *RecursiveBacktracker) Solve(g *maze.Maze, fromCell, toCell *maze.Cell, delay time.Duration, keyInput <-chan string) (*maze.Maze, error) {
	defer solvealgos.TimeTrack(a, time.Now())

	var solvePath = g.SolvePath()
	travelPath = g.TravelPath()
	startCell = fromCell

	// DFS traversal of the grid
	if r := Step(g, fromCell, toCell, solvePath, delay, keyInput); !r {
		return nil, fmt.Errorf("failed to find path through maze from %v to %v", fromCell, toCell)
	}

	g.SetPathFromTo(fromCell, toCell, solvePath)

	// stats
	a.SetSolvePath(solvePath)
	a.SetSolveSteps(travelPath.Length())
	a.SetTravelPath(travelPath)

	return g, nil
}
