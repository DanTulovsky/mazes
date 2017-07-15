package dp

import (
	"fmt"
	"mazes/maze"

	"log"

	"math"

	"mazes/utils"

	"github.com/gonum/matrix"
	"github.com/gonum/matrix/mat64"
)

const (
	North = iota
	South
	East
	West
)

type Policy struct {
	m       *mat64.Dense // the policy matrix
	actions []int
}

func reshape(m mat64.Matrix, rows, columns int) *mat64.Dense {
	mr, mc := m.Dims()
	if mr*mc != rows*columns {
		panic(matrix.ErrShape)
	}
	var r mat64.Dense
	r.Clone(m)
	raw := r.RawMatrix()
	raw.Rows = rows
	raw.Cols = columns
	raw.Stride = columns
	r.SetRawMatrix(raw)
	return &r
}

// Eval evaluates a given policy and returns a vector representing its value
//
// Recall that a policy, pi, is a mapping from each state, s in S, and action, a in A(s),
// to the probability pi(s,a) of taking action a when in state s.
// The value of a state s under policy pi is the expected return when starting in s
// and following pi thereafter.
//
// df is the discount_factor, the discount_rate: the present value of future rewards
//  df = 0; agent maximizes only current rewards
// as df approaches 1, agent becomes more farsighted
func (p *Policy) Eval(m *maze.Maze, clientID string, df float64, theta float64) (*ValueFunction, error) {
	r, _ := p.m.Dims()
	vFunction := NewValueFunction(r) // based on number of rows in matrix = number of states

	// get the cell that is the end (reward = 0)
	client, err := m.Client(clientID)
	if err != nil {
		return nil, err
	}
	endCell := m.ToCell(client)
	log.Printf("final cell is: %v", endCell)

	step := 0

	for {
		step++
		// log.Printf("step: %v", step)
		delta := 0.0 // stop when delta < theta

		// for each state, perform a full backup
		// number of states = number of rows in matrix
		for state := 0; state < r; state++ {
			v := 0.0 // expected value

			// look through all actions
			row := p.m.RowView(state)
			for action := 0; action < row.Len(); action++ {
				actionProb := row.At(action, 0)
				// log.Printf("state: %v; action: %v; v: %v", state, action, actionProb)

				// For each action, look at the possible next states

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

				// reward = -1, except at the terminal state = 0
				reward := -1.0
				var nextState int
				// figure out the next state (cell) from here given the action
				if utils.LocsSame(cell.Location(), endCell.Location()) {
					reward = 0
					nextState = state // don't move anywhere else
				} else {
					// find next cell given the action and get its state number
					switch {
					case action == North:
						if cell.North() == nil {
							nextState = state // cannot move off the grid
							break
						}
						if cell.Linked(cell.North()) {
							nextState, err = utils.StateFromLocation(m.Config().Rows, m.Config().Columns, cell.North().Location())
							if err != nil {
								return nil, err
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
								return nil, err
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
								return nil, err
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
								return nil, err
							}
						}
					}
				}

				// prob = 1; probability ???
				prob := 1.0
				// next_state = cell this move takes you to; stay in one place if can't go in that direction
				vNextState, err := vFunction.Get(nextState)
				if err != nil {
					return nil, err
				}

				// bellman equation
				v = v + actionProb*prob*(reward+df*vNextState)
			}

			// How much our value function changed (across any states)
			//delta = max(delta, np.abs(v - V[s]))
			previousVal, err := vFunction.Get(state)
			if err != nil {
				return nil, err
			}
			delta = math.Max(delta, math.Abs(v-previousVal))
			// log.Printf("delta: %v", delta)

			// store the new value for state state
			vFunction.Set(state, v)
			//log.Printf("vFunction:\n%v", vFunction.Reshape(int(m.Config().Rows), int(m.Config().Columns)))

		}

		log.Printf("delta: %v", delta)
		if delta < theta {
			break
		}
	}

	log.Printf("Steps taken: %v", step)
	return vFunction, nil

}

func NewRandomPolicy(states int, actions []int) *Policy {

	m := mat64.NewDense(states, len(actions), nil)

	setOne := func(i, j int, v float64) float64 {
		return 1.0 / float64(len(actions))
	}
	m.Apply(setOne, m)
	return &Policy{
		m:       m,
		actions: actions,
	}
}

func (p *Policy) String() string {
	r, c := p.m.Dims()
	excerpt := 0
	if r > 10 || c > 10 {
		excerpt = 3
	}
	return fmt.Sprintf("%v\n\n", mat64.Formatted(p.m, mat64.Prefix(""), mat64.Excerpt(excerpt)))
}

type ValueFunction struct {
	// should be a vector, but required to be interface for T() to work properly
	v *mat64.Vector
}

func NewValueFunction(states int) *ValueFunction {
	v := mat64.NewVector(states, nil)

	return &ValueFunction{
		v: v,
	}
}

func (vf *ValueFunction) String() string {
	r, c := vf.v.Dims()
	excerpt := 0
	if r > 10 || c > 10 {
		excerpt = 3
	}
	return fmt.Sprintf("%v\n\n", mat64.Formatted(vf.v, mat64.Prefix(""), mat64.Excerpt(excerpt)))
}

func (vf *ValueFunction) Reshape(rows, columns int) string {
	reshaped := reshape(vf.v, rows, columns)
	return fmt.Sprintf("%v\n\n", mat64.Formatted(reshaped, mat64.Prefix(""), mat64.Excerpt(0)))

}

// Set sets the value at location l to v.
func (vf *ValueFunction) Set(l int, v float64) error {
	if l > vf.v.Len() || l < 0 {
		return fmt.Errorf("invalid value for l (%v), must be between: [0,%v)", l, vf.v.Len())
	}
	vf.v.SetVec(l, v)
	return nil
}

// Get retrieves the value at index l
func (vf *ValueFunction) Get(l int) (float64, error) {
	if l > vf.v.Len() || l < 0 {
		return 0, fmt.Errorf("invalid value for l (%v), must be between: [0,%v)", l, vf.v.Len())
	}

	return vf.v.At(l, 0), nil
}
