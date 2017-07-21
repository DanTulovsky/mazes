package dp

import (
	"log"
	"mazes/maze"
)

// PolicyImprovement Algorithm. Iteratively evaluates and improves a policy
// until an optimal policy is found.
func PolicyImprovement(m *maze.Maze, clientID string, df, theta float64, actions []int) (*Policy, *ValueFunction, error) {
	numStates := int(m.Config().Columns * m.Config().Rows)
	// start with a random policy
	policy := NewRandomPolicy(numStates, actions)
	vf := NewValueFunction(numStates)

	// get the cell that is the end (reward = 0)
	client, err := m.Client(clientID)
	if err != nil {
		return nil, nil, err
	}
	endCell := m.ToCell(client).Location()

	policiesEvaluated := 0

	for {
		policiesEvaluated++

		// evaluate the current policy
		//log.Printf("evaluating current policy:\n%v", policy)
		vf, err = policy.Eval(m, clientID, df, theta, vf)
		if err != nil {
			return nil, nil, err
		}
		//log.Printf("Current Policy value function:\n%v",
		//	vf.Reshape(int(m.Config().Rows), int(m.Config().Columns)))

		// Will be set to false if we make any changes to the policy
		policyStable := true

		// For each state...
		for state := 0; state < numStates; state++ {
			// The best action we would take under the current policy
			chosenAction := policy.BestRandomActionsForState(state)

			actionValues, err := OneStepLookAhead(m, endCell, vf, df, state, len(actions))
			if err != nil {
				return nil, nil, err
			}
			bestAction := MaxInVectorIndex(actionValues)
			//log.Printf("state: %v; bestAction: %v; chosenAction: %v", state, ActionToText[bestAction], ActionToText[chosenAction])

			// Greedily update the policy
			if chosenAction != bestAction {
				policyStable = false
			}
			// create a new policy for the given state (e.g. [0, 1, 0, 0] = always go south
			newPolicyForState := make([]float64, len(actions))
			newPolicyForState[bestAction] = 1
			policy.SetState(state, newPolicyForState)
			//log.Printf("setting policy %v for state %v", newPolicyForState, state)
			//log.Printf("new current policy:\n%v", policy)
			//log.Printf("")
		}

		if policyStable {
			break
		}
		// log.Printf("policy:\n%v\nvf:\n%v", policy, vf.Reshape(int(m.Config().Rows), int(m.Config().Columns)))
	}

	log.Printf("policies evaluated: %v", policiesEvaluated)
	return policy, vf, nil
}
