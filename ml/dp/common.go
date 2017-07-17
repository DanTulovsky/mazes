package dp

import (
	"log"
	"math"
	"mazes/maze"
	"mazes/utils"

	"fmt"

	"github.com/gonum/matrix/mat64"
)

const (
	North = iota
	South
	East
	West
	None // no movement is best
)

var (
	DefaultActions = []int{North, South, East, West, None}
	ActionToText   = map[int]string{
		North: "north",
		South: "south",
		East:  "east",
		West:  "west",
		None:  "none",
	}
)

// MaxInVectorIndex returns the position of the max element in the vector
// ties are broken arbitrarily
func MaxInVectorIndex(v *mat64.Vector) int {

	max := math.Inf(-1)
	var best []int

	for x := 0; x < v.Len(); x++ {
		value := v.At(x, 0)
		if value >= max {
			if value == max {
				best = append(best, x)
			} else {
				best = []int{x}
			}
			max = value
		}
	}
	return best[utils.Random(0, len(best))]
}

func probabilityForStateAction(m *maze.Maze, state, nextState, a int) (float64, error) {
	p := 0.0

	// probability of moving from state -> nextState via a
	// if cells linked = 1
	// otherwise = 0
	locCellFrom, err := utils.LocationFromState(m.Config().Rows, m.Config().Columns, int64(state))
	if err != nil {
		return 0.0, err
	}
	cellFrom, err := m.Cell(locCellFrom.GetX(), locCellFrom.GetY(), locCellFrom.GetZ())
	if err != nil {
		return 0.0, err
	}

	locCellTo, err := utils.LocationFromState(m.Config().Rows, m.Config().Columns, int64(nextState))
	if err != nil {
		return 0.0, err
	}
	cellTo, err := m.Cell(locCellTo.GetX(), locCellTo.GetY(), locCellTo.GetZ())
	if err != nil {
		return 0.0, err
	}

	switch {
	case ActionToText[a] == "north":
		if cellFrom.Linked(cellFrom.North()) && utils.LocsSame(locCellTo, cellFrom.North().Location()) {
			p = 1.0
		}
	case ActionToText[a] == "south":
		if cellFrom.South() != nil && utils.LocsSame(locCellTo, cellFrom.South().Location()) {
			p = 1.0
		}
	case ActionToText[a] == "east":
		if cellFrom.East() != nil && utils.LocsSame(locCellTo, cellFrom.East().Location()) {
			p = 1.0
		}
	case ActionToText[a] == "west":
		if cellFrom.West() != nil && utils.LocsSame(locCellTo, cellFrom.West().Location()) {
			p = 1.0
		}
	case ActionToText[a] == "none":
		p = 1.0
	default:
		return 0.0, fmt.Errorf("probabilityForStateAction: invalid action: %v", ActionToText[a])
	}

	log.Printf("Prob: %v -> %v (via %v): %v (%v)", cellFrom, cellTo, ActionToText[a], cellFrom.Linked(cellTo), p)

	return p, nil
}

// OneStepLookAhead returns a vector of expected values for each action
func OneStepLookAhead(m *maze.Maze, endCell *maze.Cell, vf *ValueFunction, df float64, state, numActions int) (*mat64.Vector, error) {

	actionValues := mat64.NewVector(numActions, nil)

	// Find the best action by one-step lookahead, ties resolved arbitrarily
	for a := 0; a < numActions; a++ {
		nextState, reward, err := NextState(m, endCell, state, a)
		if err != nil {
			return nil, err
		}

		// Probablity of transitioning from state -> nextState given action a.
		prob, err := probabilityForStateAction(m, state, nextState, a)
		if err != nil {
			return nil, err
		}
		//log.Printf("currentState: %v; action: %v; nextState: %v; reward: %v", state, ActionToText[a], nextState, reward)

		vNextState, err := vf.Get(nextState)
		if err != nil {
			return nil, err
		}

		// current value
		v := actionValues.At(a, 0)
		v = v + prob*(reward+df*vNextState)
		actionValues.SetVec(a, v)
	}

	return actionValues, nil
}

// NextState returns the next state (as int) given the current state and action
// returns nextState, reward, error
func NextState(m *maze.Maze, endCell *maze.Cell, state, action int) (int, float64, error) {
	var nextState int
	var reward float64
	// For each action, look at the possible next states
	cell, err := CellFromState(m, state)
	if err != nil {
		return 0, 0, err
	}

	// figure out the next state (cell) from here given the action
	if utils.LocsSame(cell.Location(), endCell.Location()) {
		reward = 0
		nextState = state // don't move anywhere else
	} else {
		reward = -1
		// find next cell given the action and get its state number
		switch {
		case action == None:
			nextState = state // no movement
		case action == North:
			if cell.North() == nil {
				nextState = state // cannot move off the grid
				break
			}
			if cell.Linked(cell.North()) {
				nextState, err = utils.StateFromLocation(m.Config().Rows, m.Config().Columns, cell.North().Location())
				if err != nil {
					return 0, 0, err
				}
			}
		case action == South:
			if cell.South() == nil {
				nextState = state // cannot move off the grid
				break
			}
			if cell.Linked(cell.South()) {
				nextState, err = utils.StateFromLocation(m.Config().Rows, m.Config().Columns, cell.South().Location())
				if err != nil {
					return 0, 0, err
				}
			}
		case action == East:
			if cell.East() == nil {
				nextState = state // cannot move off the grid
				break
			}
			if cell.Linked(cell.East()) {
				nextState, err = utils.StateFromLocation(m.Config().Rows, m.Config().Columns, cell.East().Location())
				if err != nil {
					return 0, 0, err
				}
			}
		case action == West:
			if cell.West() == nil {
				nextState = state // cannot move off the grid
				break
			}
			if cell.Linked(cell.West()) {
				nextState, err = utils.StateFromLocation(m.Config().Rows, m.Config().Columns, cell.West().Location())
				if err != nil {
					return 0, 0, err
				}
			}
		}
	}
	return nextState, reward, nil
}

// CellFromState returns the cell given the state number
func CellFromState(m *maze.Maze, state int) (*maze.Cell, error) {
	// get cell from state; the state is simply an integer that counts the cells in the maze
	// starting from the top left and going row by row
	l, err := utils.LocationFromState(m.Config().Rows, m.Config().Columns, int64(state))
	if err != nil {
		return nil, err
	}
	cell, err := m.Cell(l.X, l.Y, l.Z)
	if err != nil {
		return nil, fmt.Errorf("failed to find cell at %v: %v (state=%v)", l, err, state)
	}
	return cell, nil
}
