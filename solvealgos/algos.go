// Package solvealgos defines the interface for solver algorithms
package solvealgos

import (
	"errors"
	"mazes/grid"
	"time"
)

type Algorithmer interface {
	SolvePath() []*grid.Cell
	SolveTime() time.Duration
	SetSolvePath(p []*grid.Cell)
	SetSolveTime(t time.Duration)
	Solve(*grid.Grid, *grid.Cell, *grid.Cell) (*grid.Grid, error)
}

type Common struct {
	solvePath  []*grid.Cell  // path of the final solution
	solveSteps int           // how many cell visits it tooks (including duplicates)
	solveTime  time.Duration // how long the last solve time took
}

// Solve should write the path of the solution to the grid
func (a *Common) Solve(g *grid.Grid, fromCell, toCell *grid.Cell) (*grid.Grid, error) {
	return nil, errors.New("Solve() not implemented")
}

func TimeTrack(a Algorithmer, start time.Time) {
	a.SetSolveTime(time.Since(start))
}

func (a *Common) SolveTime() time.Duration {
	return a.solveTime
}

func (a *Common) SetSolveTime(t time.Duration) {
	a.solveTime = t
}

func (a *Common) SolvePath() []*grid.Cell {
	return a.solvePath
}

func (a *Common) SetSolvePath(p []*grid.Cell) {
	a.solvePath = p
}
