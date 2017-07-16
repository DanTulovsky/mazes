package dp

import (
	"log"
	"mazes/maze"

	"github.com/gonum/matrix/mat64"
)

// PolicyImprovement Algorithm. Iteratively evaluates and improves a policy
// until an optimal policy is found.
func PolicyImprovement(m *maze.Maze, clientID string, df, theta float64, actions []int) (*Policy, *ValueFunction, error) {
	// start with a random policy
	policy := NewRandomPolicy(int(m.Config().Rows*m.Config().Columns), actions)
	vf := NewValueFunction(int(m.Config().Rows * m.Config().Columns))
	numStates := int(m.Config().Columns * m.Config().Rows)

	// get the cell that is the end (reward = 0)
	client, err := m.Client(clientID)
	if err != nil {
		return nil, nil, err
	}
	endCell := m.ToCell(client)

	step := 0

	for {
		step++

		// evaluate the current policy
		log.Printf("evaluating current policy:\n%v", policy)
		vf, err = policy.Eval(m, clientID, df, theta)
		if err != nil {
			return nil, nil, err
		}
		log.Printf("Current Policy value function:\n%v",
			vf.Reshape(int(m.Config().Rows), int(m.Config().Columns)))

		// Will be set to false if we make any changes to the policy
		policyStable := true

		// For each state...
		for state := 0; state < numStates; state++ {
			// The best action we would take under the current policy
			chosenAction := policy.BestRandomActionsForState(state)

			actionValues := mat64.NewVector(len(actions), nil)
			// Find the best action by one-step lookahead, ties resolved arbitrarily
			for a := 0; a < len(actions); a++ {
				prob := 1.0
				nextState, reward, err := policy.NextState(m, endCell, state, a)
				if err != nil {
					return nil, nil, err
				}
				log.Printf("nextState: %v; reward: %v", nextState, reward)

				vNextState, err := vf.Get(nextState)
				if err != nil {
					return nil, nil, err
				}
				log.Printf("vNextState: %v", vNextState)

				// current value
				v := actionValues.At(a, 0)
				v = v + prob*(reward+df*vNextState)
				actionValues.SetVec(a, v)
				log.Printf("setting action %v to %v", a, v)
			}
			bestAction := MaxInVector(actionValues)

			// Greedily update the policy
			if chosenAction != bestAction {
				policyStable = false
				// create a new policy for the given state (e.g. [0, 1, 0, 0] = always go south
				newPolicyForState := make([]float64, len(actions))
				newPolicyForState[bestAction] = 1
				policy.SetState(state, newPolicyForState)
			}
		}

		if policyStable {
			return policy, vf, nil
		}
	}

	log.Printf("steps: %v", step)
	return policy, vf, nil
}
