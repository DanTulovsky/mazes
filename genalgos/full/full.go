// Package full creates a grid with all walls in place
package full

import (
	"time"

	"gogs.wetsnow.com/dant/mazes/genalgos"
	"gogs.wetsnow.com/dant/mazes/maze"

	"fmt"

	"github.com/tevino/abool"
)

type Full struct {
	genalgos.Common
}

// Apply doesn't do anything, the grid is empty.
func (a *Full) Apply(m *maze.Maze, delay time.Duration, generating *abool.AtomicBool) error {

	defer genalgos.TimeTrack(m, time.Now())

	for currentCell := range m.Cells() {
		if !generating.IsSet() {
			return fmt.Errorf("stop requested")
		}

		time.Sleep(delay) // animation delay
		m.SetGenCurrentLocation(currentCell)

		//for _, n := range currentCell.Neighbors() {
		//	m.Link(currentCell, n)
		//}
	}

	a.Cleanup(m)
	return nil
}

func (a *Full) CheckGrid(m *maze.Maze) error {
	return nil
}
