package recursive_division

import (
	"log"
	"mazes/genalgos"
	"mazes/maze"
	"mazes/utils"
	"time"
)

type RecursiveDivision struct {
	genalgos.Common
}

// initMaze initializes the maze by linking all cells together to create one large space
func initMaze(m *maze.Maze) {
	// weaving returns remote neighbors, this doesn't work here
	if m.Config().AllowWeaving {
		log.Print("disabling weaving, it's not supported for this algorithm")
		m.Config().AllowWeaving = false
	}

	for c := range m.Cells() {
		for _, n := range c.Neighbors() {
			// Does double the work by linking all cells twice
			m.Link(c, n)
		}
	}

}

func divide(m *maze.Maze, row, column, height, width int, delay time.Duration) error {

	if height <= 1 || width <= 1 {
		return nil
	}

	if height > width {
		divideHorizontally(m, row, column, height, width, delay)
	} else {
		divideVertically(m, row, column, height, width, delay)
	}

	return nil
}

func divideHorizontally(m *maze.Maze, row, column, height, width int, delay time.Duration) {

	divideSouthOf := utils.Random(0, height-1)
	passageAt := utils.Random(0, width)

	for x := 0; x < width; x++ {
		time.Sleep(delay) // animation delay

		if x == passageAt {
			continue // keep this passage open
		}

		if cell, err := m.Cell(column+x, row+divideSouthOf, 0); err != nil {
			log.Fatalf("failed to get cell at [%v, %v, %v]", column+x, row+divideSouthOf, 0)
		} else {
			if cell.South() != nil {
				cell.UnLink(cell.South())
			}
		}
	}

	divide(m, row, column, divideSouthOf+1, width, delay)
	divide(m, row+divideSouthOf+1, column, height-divideSouthOf-1, width, delay)
}

func divideVertically(m *maze.Maze, row, column, height, width int, delay time.Duration) {

	divideEastOf := utils.Random(0, width-1)
	passageAt := utils.Random(0, height)

	for y := 0; y < height; y++ {
		time.Sleep(delay) // animation delay

		if y == passageAt {
			continue // keep this passage open
		}

		if cell, err := m.Cell(column+divideEastOf, row+y, 0); err != nil {
			log.Fatalf("failed to get cell at [%v, %v, %v]", column+divideEastOf, row+y, 0)
		} else {
			if cell.East() != nil {
				cell.UnLink(cell.East())
			}
		}
	}

	divide(m, row, column, height, divideEastOf+1, delay)
	divide(m, row, column+divideEastOf+1, height, width-divideEastOf-1, delay)
}

// Apply applies the binary tree algorithm to generate the maze.
func (a *RecursiveDivision) Apply(m *maze.Maze, delay time.Duration) error {

	defer genalgos.TimeTrack(m, time.Now())

	// links all cells together
	initMaze(m)

	width, height := m.Dimensions()
	divide(m, 0, 0, height, width, delay)

	a.Cleanup(m)
	return nil
}
