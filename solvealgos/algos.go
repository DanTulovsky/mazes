// Package solvealgos defines the interface for solver algorithms
package solvealgos

import (
	"errors"
	"mazes/grid"
	"time"
)

type Algorithmer interface {
	SolvePath() []*grid.Cell // final path
	SolveSteps() int
	SolveTime() time.Duration
	SetSolvePath(p []*grid.Cell)
	SetSolveSteps(s int)
	SetSolveTime(t time.Duration)
	Solve(*grid.Grid, *grid.Cell, *grid.Cell) (*grid.Grid, error)
	TravelPath() []*grid.Cell // all the cells travelled
}

type Common struct {
	solvePath  []*grid.Cell  // path of the final solution
	solveSteps int           // how many cell visits it tooks (including duplicates)
	solveTime  time.Duration // how long the last solve time took
	travelPath []*grid.Cell  // all the cells visited in order
}

// Solve should write the path of the solution to the grid
func (a *Common) Solve(g *grid.Grid, fromCell, toCell *grid.Cell) (*grid.Grid, error) {
	return nil, errors.New("Solve() not implemented")
}

// TimeTrack tracks sets the time it took for the algorithm to run
func TimeTrack(a Algorithmer, start time.Time) {
	a.SetSolveTime(time.Since(start))
}

// SolveTime returns the time it took to solve the maze
func (a *Common) SolveTime() time.Duration {
	return a.solveTime
}

// SetSolveTime sets solveTime
func (a *Common) SetSolveTime(t time.Duration) {
	a.solveTime = t
}

// SolvePath returns the path for the solution
func (a *Common) SolvePath() []*grid.Cell {
	return a.solvePath
}

// SetSolvePath sets the solvePath
func (a *Common) SetSolvePath(p []*grid.Cell) {
	a.solvePath = p
}

// SolveSteps returns the number of steps (visits to cells) it took to solve the maze
func (a *Common) SolveSteps() int {
	return a.solveSteps
}

// SetSolveSteps sets the solveSteps
func (a *Common) SetSolveSteps(s int) {
	a.solveSteps = s
}

// TravelPath returns the entire path traveled (often the same as the solution path)
func (a *Common) TravelPath() []*grid.Cell {
	return a.travelPath
}

// SetTravelPath sets the solvePath
func (a *Common) SetTravelPath(p []*grid.Cell) {
	a.travelPath = p
}
