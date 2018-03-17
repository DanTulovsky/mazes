package rules

import (
	"mazes/colors"
	"mazes/maze"
)

// Classic implements the classic game of life rules
func Classic(m *maze.Maze) *maze.Maze {

	for c := range m.Cells() {
		liveNeighbors := 0
		for _, n := range c.AllNeighbors() {
			if n.BGColor() == colors.GetColor("black") {
				liveNeighbors++
			}
		}

		if liveNeighbors < 2 {
			// die, lonely
			defer c.SetBGColor(colors.GetColor("white"))
			continue
		}

		if liveNeighbors > 3 {
			// die, overcrowded
			defer c.SetBGColor(colors.GetColor("white"))
			continue
		}

		if liveNeighbors == 3 && c.BGColor() == colors.GetColor("white") {
			// Dead cell with 3 live neighbors becomes alive
			defer c.SetBGColor(colors.GetColor("black"))

		}
	}
	return m
}
