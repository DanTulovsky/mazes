package states

import (
	"github.com/DanTulovsky/mazes/automata/rules"
	"github.com/DanTulovsky/mazes/colors"
	"github.com/DanTulovsky/mazes/maze"
	"github.com/DanTulovsky/mazes/utils"
)

// Empty doesn't set any initial state
func Empty(m *maze.Maze) *maze.Maze {
	return m
}

// Random sets a random initial state for game of life
func Random(m *maze.Maze) *maze.Maze {
	for c := range m.Cells() {
		if utils.Random(0, 2) == 0 {
			c.SetBGColor(colors.GetColor(rules.AliveColor))
		}
	}
	return m
}

// Concrete sets a concrete initial state for game of life
func Concrete(m *maze.Maze) *maze.Maze {
	liveCells := []*maze.Cell{
		// block
		m.CellBeSure(1, 1, 0),
		m.CellBeSure(1, 2, 0),
		m.CellBeSure(2, 1, 0),
		m.CellBeSure(2, 2, 0),
		// line
		// m.CellBeSure(5, 4, 0),
		// m.CellBeSure(5, 5, 0),
		// m.CellBeSure(5, 6, 0),
		// // two diagonal blocks
		// m.CellBeSure(8, 1, 0),
		// m.CellBeSure(8, 2, 0),
		// m.CellBeSure(9, 1, 0),
		// m.CellBeSure(9, 2, 0),
		// m.CellBeSure(10, 3, 0),
		// m.CellBeSure(10, 4, 0),
		// m.CellBeSure(11, 3, 0),
		// m.CellBeSure(11, 4, 0),
	}

	for _, c := range liveCells {
		c.SetBGColor(colors.GetColor(rules.AliveColor))
	}
	return m
}
