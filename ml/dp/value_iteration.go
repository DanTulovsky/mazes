package dp

import (
	"log"
	"math"
	"github.com/DanTulovsky/mazes/maze"
	"github.com/DanTulovsky/mazes/ml"

	"github.com/gonum/matrix/mat64"
)

func ValueIteration(m *maze.Maze, clientID string, df, theta float64, actions []int) (*ml.Policy, *ml.ValueFunction, error) {

	// Used to construct value function V
	numStates := int(m.Config().Columns * m.Config().Rows)

	// get the cell that is the end (reward = 0)
	client, err := m.Client(clientID)
	if err != nil {
		return nil, nil, err
	}
	endCell := m.ToCell(client).Location()

	// new random value function
	vf := ml.NewValueFunction(numStates)

	vfEvaluated := 0

	// Do value iteration until delta < theta
	for {
		vfEvaluated++
		delta := 0.0

		// For each state...
		for state := 0; state < numStates; state++ {

			actionValues, err := ml.OneStepLookAhead(m, endCell, vf, df, state, len(actions))
			if err != nil {
				return nil, nil, err
			}
			bestActionValue := mat64.Max(actionValues)
			// log.Printf("best: %v, actionValues: %v", bestActionValue, mat64.Formatted(actionValues.T(), mat64.Prefix(""), mat64.Excerpt(0)))

			// How much our value function changed (across any states)
			previousVal, err := vf.Get(state)
			if err != nil {
				return nil, nil, err
			}
			delta = math.Max(delta, math.Abs(bestActionValue-previousVal))

			// Update the value function
			vf.Set(state, bestActionValue)
		}

		// log.Printf("delta: %v", delta)
		if delta < theta {
			break
		}
	}

	log.Printf("value functions evaluated: %v", vfEvaluated)

	// Build policy based on value function
	policy, err := ml.NewPolicyFromValueFunction(m, endCell, vf, df, numStates, actions)
	if err != nil {
		return nil, nil, err
	}

	return policy, vf, nil
}
