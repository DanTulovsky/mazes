package dp

import (
	"math"
	"gogs.wetsnow.com/dant/mazes/maze"
	"gogs.wetsnow.com/dant/mazes/ml"
)

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
func Evaluate(p *ml.Policy, m *maze.Maze, clientID string, df float64, theta float64, vf *ml.ValueFunction) (*ml.ValueFunction, error) {
	numStates := int(m.Config().Columns * m.Config().Rows)
	if vf == nil {
		vf = ml.NewValueFunction(numStates) // based on number of rows in matrix = number of states
	}

	// get the cell that is the end (reward = 0)
	client, err := m.Client(clientID)
	if err != nil {
		return nil, err
	}
	endCell := m.ToCell(client).Location()

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
			actions := p.M.RowView(state)
			for action := 0; action < actions.Len(); action++ {
				// probability of taking action a at state s under stochastic policy pi
				actionProb := actions.At(action, 0)
				// log.Printf("state: %v; action: %v; v: %v", state, action, actionProb)

				// reward = -1, except at the terminal state = 0
				// expected immediate reward on transition from s to s' under action a
				reward := -1.0
				nextState, reward, valid, err := ml.NextState(m, endCell, state, action)
				if err != nil {
					return nil, err
				}
				if !valid {
					continue // do not include actions that are not possible from this state
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
