package dp

import (
	"fmt"
	"mazes/maze"

	"math"

	"mazes/utils"

	"github.com/gonum/matrix"
	"github.com/gonum/matrix/mat64"
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

// NewZeroPolicy returns a policy that gives each action weight of 0
func NewZeroPolicy(numStates int, actions []int) *Policy {
	return &Policy{
		m:       mat64.NewDense(numStates, len(actions), nil),
		actions: actions,
	}
}

// NewRandomPolicy returns a policy that gives each action the same weight
func NewRandomPolicy(numStates int, actions []int) *Policy {

	m := mat64.NewDense(numStates, len(actions), nil)

	setOne := func(i, j int, v float64) float64 {
		return 1.0 / float64(len(actions))
	}
	m.Apply(setOne, m)
	return &Policy{
		m:       m,
		actions: actions,
	}
}

// NewPolicyFromValuFunction returns a policy based on the provided value function
func NewPolicyFromValuFunction(m *maze.Maze, endCell *maze.Cell, vf *ValueFunction, df float64, numStates int, actions []int) (*Policy, error) {
	policy := NewZeroPolicy(numStates, actions)

	for state := 0; state < numStates; state++ {
		// One step lookahead to find the best action for this state
		actionValues, err := OneStepLookAhead(m, endCell, vf, df, state, len(actions))
		if err != nil {
			return nil, err
		}

		bestAction := MaxInVectorIndex(actionValues)
		// Always take the best action
		policy.SetStateAction(state, bestAction, 1.0)
	}
	return policy, nil
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
// vf is the value function to use, if nil, a new, all zero, one is used
func (p *Policy) Eval(m *maze.Maze, clientID string, df float64, theta float64, vf *ValueFunction) (*ValueFunction, error) {
	numStates := int(m.Config().Columns * m.Config().Rows)
	if vf == nil {
		vf = NewValueFunction(numStates) // based on number of rows in matrix = number of states
	}

	// get the cell that is the end (reward = 0)
	client, err := m.Client(clientID)
	if err != nil {
		return nil, err
	}
	endCell := m.ToCell(client)

	step := 0

	for {
		step++
		// log.Printf("step: %v", step)
		delta := 0.0 // stop when delta < theta

		// for each state, perform a full backup
		// number of states = number of rows in matrix
		for state := 0; state < numStates; state++ {
			v := 0.0 // expected value

			// look through all actions
			actions := p.m.RowView(state)
			for action := 0; action < actions.Len(); action++ {
				// probability of taking action a at state s under stochastic policy pi
				actionProb := actions.At(action, 0)
				// log.Printf("state: %v; action: %v; v: %v", state, action, actionProb)

				// reward = -1, except at the terminal state = 0
				// expected immediate reward on transition from s to s' under action a
				reward := -1.0
				nextState, reward, err := NextState(m, endCell, state, action)
				if err != nil {
					return nil, err
				}

				// prob = 1; probability of transition from s to s' under action a (always 100%)
				prob := 1.0
				// next_state = cell this move takes you to; stay in one place if can't go in that direction
				vNextState, err := vf.Get(nextState)
				if err != nil {
					return nil, err
				}

				// bellman equation
				v = v + actionProb*prob*(reward+df*vNextState)
			}

			// How much our value function changed (across any states)
			//delta = max(delta, np.abs(v - V[s]))
			previousVal, err := vf.Get(state)
			if err != nil {
				return nil, err
			}
			delta = math.Max(delta, math.Abs(v-previousVal))
			// log.Printf("delta: %v", delta)

			// store the new value for state state, in-place dynamic programming, start using new value right away
			vf.Set(state, v)
			// log.Printf("vf:\n%v", vf.Reshape(int(m.Config().Rows), int(m.Config().Columns)))

		}

		//log.Printf("delta: %v", delta)

		if delta < theta {
			break
		}
	}

	// log.Printf("Steps taken: %v", step)
	return vf, nil

}

func (p *Policy) SetState(state int, values []float64) {
	p.m.SetRow(state, values)

}

func (p *Policy) SetStateAction(state, action int, value float64) {
	p.m.Set(state, action, value)
}

func (p *Policy) ActionsForState(s int) *mat64.Vector {
	return p.m.RowView(s)
}

func (p *Policy) BestRandomActionsForState(s int) int {
	actions := p.m.RowView(s)
	max := math.Inf(-1)
	var bestActions []int

	for x := 0; x < actions.Len(); x++ {
		a := actions.At(x, 0)
		if a >= max {
			if a == max {
				bestActions = append(bestActions, x)
			} else {
				bestActions = []int{x}
			}
			max = a
		}
	}
	return bestActions[utils.Random(0, len(bestActions))]
}

func (p *Policy) String() string {
	r, c := p.m.Dims()
	excerpt := 0
	if r > 10 || c > 10 {
		excerpt = 5
	}
	return fmt.Sprintf("%v\n\n", mat64.Formatted(p.m, mat64.Prefix(""), mat64.Excerpt(excerpt)))
}
