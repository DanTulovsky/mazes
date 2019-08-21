package rules

import (
	"log"

	"github.com/DanTulovsky/mazes/colors"
	"github.com/DanTulovsky/mazes/maze"
	"github.com/DanTulovsky/mazes/utils"
)

const (
	// AliveColor is the color for the live squares.
	AliveColor = "blue"
	// DeadColor is the color for the dead squares.
	DeadColor = "white"
)

// isAlive returns true if the cell is alive
func isAlive(c *maze.Cell) bool {
	return c.BGColor() == colors.GetColor(AliveColor)
}

func isDead(c *maze.Cell) bool {
	return c.BGColor() == colors.GetColor(DeadColor)
}

// Revive sets the cell to be alive
func Revive(c *maze.Cell) {
	c.SetBGColor(colors.GetColor(AliveColor))
}

// Kill kills the square
func Kill(c *maze.Cell) {
	c.SetBGColor(colors.GetColor(DeadColor))
	c.SetDistance(0)
}

// Classic implements the classic game of life rules
func Classic(m *maze.Maze) *maze.Maze {

	for c := range m.Cells() {
		liveNeighbors := 0

		// count number of live neighbors
		for _, n := range c.AllNeighbors() {
			if isAlive(n) {
				liveNeighbors++
			}
		}

		// die, lonely
		if liveNeighbors < 2 {
			defer Kill(c)
			continue
		}

		// die, overcrowded
		if liveNeighbors > 3 {
			defer Kill(c)
			continue
		}

		// Dead cell with 3 live neighbors becomes alive
		if liveNeighbors == 3 && isDead(c) {
			defer Revive(c)

		}
	}
	return m
}

// Play1 is liek classic, but cells with nothing around them have a small chance of coming to life.
func Play1(m *maze.Maze) *maze.Maze {

	for c := range m.Cells() {
		liveNeighbors := 0
		for _, n := range c.AllNeighbors() {
			if isAlive(n) {
				liveNeighbors++
			}
		}

		// Dead cell with 0 live neighbors has a small chance of becoming alive
		if liveNeighbors == 0 && isDead(c) {
			if utils.Random(1, 1001) > 999 {
				defer Revive(c)
			}
			continue
		}

		if liveNeighbors < 2 {
			// die, lonely
			defer Kill(c)
			continue
		}

		if liveNeighbors > 3 {
			// die, overcrowded
			defer Kill(c)
			continue
		}

		// Dead cell with 3 live neighbors becomes alive
		if liveNeighbors == 3 && isDead(c) {
			defer Revive(c)
			continue
		}

	}
	return m
}

// Play2 implements some rules.
func Play2(m *maze.Maze) *maze.Maze {
	for c := range m.Cells() {
		if isAlive(c) {
			// incremenet the distance travelled, this is the strength
			c.IncDistance()
			log.Printf("%v (d=%v)", c, c.Distance())

			// pick random neighbor to move to
			rnd := c.RandomAllNeighbor()

			// "move" by killing self and enabling neighbor
			if isDead(rnd) {
				defer Revive(rnd)
				// transfer distance travelled
				defer rnd.SetDistance(c.Distance())
				defer Kill(c)
				continue
			}

			if isAlive(rnd) {
				if c.Distance() > rnd.Distance() {
					// consume the weeker neighbor
					defer rnd.SetDistance(c.Distance())
					defer Kill(c)
					continue
				}
			}
		}
	}
	return m
}
