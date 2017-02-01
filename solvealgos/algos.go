// Package solvealgos defines the interface for solver algorithms
package solvealgos

import (
	"errors"
	"mazes/grid"
	"time"
)

type Algorithmer interface {
	LastSolveTime() time.Duration
	SetSolveTime(t time.Duration)
	Solve(*grid.Grid, *grid.Cell, *grid.Cell) (*grid.Grid, error)
}

type Common struct {
	solveTime time.Duration // how long the last solve time took
}

// Solve should write the path of the solution to the grid
func (a *Common) Solve(g *grid.Grid, fromCell, toCell *grid.Cell) (*grid.Grid, error) {
	return nil, errors.New("Solve() not implemented")
}

func TimeTrack(a Algorithmer, start time.Time) {
	a.SetSolveTime(time.Since(start))
}

func (a *Common) LastSolveTime() time.Duration {
	return a.solveTime
}

func (a *Common) SetSolveTime(t time.Duration) {
	a.solveTime = t
}
